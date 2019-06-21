/*
 * Main contains the logic to create routing based on Mux and their respective handling
 * Supports
 *     Get All Config (Not implemented for now)
 *     Create Environment
 *     Get Environment based on tag
 *     Get Environment
 *     Stop a Container based on tag
 *     Delete a Container based on tag
 *
 * Hard coded to support only Mongo and Redis images
 *
 * API version: 1.0.0
 * Contact: Arun K, Vibhore
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"regexp"
	"net/http"
	"os"
	//"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"webserver/db"
	"webserver/dockercontainer"
	"webserver/logging"
	"webserver/util"
)

var (
	baseImageRegistry = "docker.io/library/"
	ctx = context.Background()
	mongoPortID string
)


func newRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", getallconfighandler).Methods("GET")
	r.HandleFunc("/set/createenv", createenvhandler).Methods("POST")
	r.HandleFunc("/get/getenv/{tag}", getenvbytaghandler).Methods("GET")
	r.HandleFunc("/get/getenv", getenvhandler).Methods("GET")
	r.HandleFunc("/update/stop/{tag}", stophandler).Methods("POST")
	r.HandleFunc("/delete/container/{tag}", delhandler).Methods("DELETE","POST")
	return r
}


//initResp is the initial response struct
type initResp struct {
	Status    string `json:"status"`
	RequestID string `json:"requestid"`
}

//requestData is the request struct
type postRequestBody struct {
	Name       string   `json:"name"`
	Containers []string `json:"containers"`
}

func main() {
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	logging.Info.Println("Initializing router")
	r := newRouter()

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logging.Info.Println("Initialize test bed meta collection")
	db.InitTestBedMetaCollection(ctx)

	logging.Info.Println("Starting Server")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

// Handler for /getenv call
func getenvhandler(w http.ResponseWriter, r *http.Request) {
	containersList := dockercontainer.ListContainers(ctx)
	fmt.Fprintf(w, "Listing all the containers : \n")

        for _, container := range containersList {
		logging.Info.Println(container.Names[0] + " " + container.ID[:10] )
		fmt.Fprintf(w, container.Names[0] + " " + container.ID[:10] + " \n")
        }
}

// Handler for /getenv/<container-id> call read from Mongo
func getenvbytaghandler(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        testbedID := vars["tag"]

	if testbedID != "" {
		testbedInfo, err := db.GetTestBedFromID(ctx, testbedID)
		if err != nil {
			logging.Error.Println("Error observed while fetching testbedInfo.")
		} else {
			logging.Info.Println(testbedInfo)
			// todo - Add logic to print detailed test bed info
			fmt.Fprintf(w, "%v", testbedInfo)
			//fmt.Println(w, testbedInfo.ID)
			//fmt.Println(w, testbedInfo.Name)
		}
	} else {
		logging.Error.Println("No tag information provided.")
	}
}


/* 
  Handler for /createenv call

  This call will be used for capturing the json request body which will be then written to a file. 
  Once we have data in the file, a docker image pull and container creation will begin accordingly. 
*/
func createenvhandler(w http.ResponseWriter, r *http.Request) {
	post :=  postRequestBody{}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	defer r.Body.Close()
	testbed := db.NewTestBed()

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		s := err.Error()
		logging.Error.Println(s)
	}

        requestBody, _ := ioutil.ReadAll(r.Body)

	logging.Info.Println(requestBody)

        err1 := json.Unmarshal(requestBody, &post)
        if err1 != nil{
                logging.Error.Println(err1)
        }

	testbed.Name = post.Name
	for _, cnt := range post.Containers {
		testbed.Container = append(testbed.Container, db.ContainerProp{Image: cnt, CID: "0", IP: "0.0.0.0"})
	}

	insertResult, err := db.InsertTestBed(context.TODO(), testbed)
	if err != nil {
		logging.Error.Println(err)
	}

	logging.Info.Println("Created testbed document: ", insertResult.InsertedID)

	tbID := insertResult.InsertedID.(string)

	go pullDockerImageAndCreateContainer(tbID, post.Containers)

	rsp := initResp{Status: "pending", RequestID: tbID}
	json.NewEncoder(w).Encode(rsp)
}

/*
  pullDockerImageAndCreateContainer is used to pull docker images and create container
  Pulling docker images is a goroutine based implementation.
*/
func pullDockerImageAndCreateContainer(tbid string, containers []string) {
	var images []string

	logging.Info.Println("Initializing wait group")
	var wg sync.WaitGroup

	for _, container := range containers {
		if container == "mongo" || container == "redis" {
			imageName := baseImageRegistry + strings.ToLower(container)
			logging.Info.Println( "Image name is " + imageName )
			images = append(images, imageName)
			wg.Add(1)
	                go dockercontainer.PullDockerImages(ctx, imageName, &wg)
		}
	}

	logging.Info.Println("Images list is : ", images)

	wg.Wait()

	tag := tbid


	_, err := db.UpdateTestBedStatus(context.TODO(), tbid, "In-progress")
        if err != nil {
                logging.Error.Println(err)
        }

	for _, image := range images {
		sImage := strings.Split(image, "/")
		image = sImage[len(sImage)-1]
		//fmt.Fprintf(w, image)
		if image == "mongo" || image == "redis" {
			port, err := util.GetFreePort()
			resp, containerPortString := dockercontainer.CreateDockerContainer(ctx, image, tag, port)
			containers = append(containers, resp.ID)
			db.AddPortToMeta(context.TODO(), port)

			dockercontainer.StartContainer(ctx, resp)
			inspectData := dockercontainer.InspectContainer(ctx, resp.ID)
			logging.Info.Println("IP Address for container : ", inspectData.NetworkSettings.IPAddress)
			logging.Info.Println("Port map for container : ", inspectData.NetworkSettings.Ports)

			logging.Info.Println("Building container : " + image)
			containerIP := inspectData.NetworkSettings.IPAddress

			hostport := inspectData.NetworkSettings.Ports[containerPortString][0].HostPort
			hport, err1 := strconv.Atoi(hostport)
			if err1 != nil {
				logging.Error.Println(err)
			}
			logging.Info.Println("Host port value is ", hostport)

			_, err = db.UpdateContainerProperty(context.TODO(), tbid, image, "ip", containerIP)
			if err != nil {
				logging.Error.Println(err)
			}
			_, err = db.UpdateContainerProperty(context.TODO(), tbid, image, "svc_port", hport)
			if err != nil {
				logging.Error.Println(err)
			}
			_, err = db.UpdateContainerProperty(context.TODO(), tbid, image, "rest_port", 7010)
			if err != nil {
				logging.Error.Println(err)
			}
			logging.Info.Println("Done building container: " + image)
		}
	}

	_, err = db.UpdateTestBedStatus(context.TODO(), tbid, "Completed")
        if err != nil {
                logging.Error.Println(err)
        }
}


// Handler for / call
func getallconfighandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Nothing to do here as of now.\n")
}


// Handler for /stop request
func stophandler(w http.ResponseWriter, r *http.Request) {
	var stoppedContainer []string

	vars := mux.Vars(r)
	tag := vars["tag"]

	containersList := dockercontainer.ListContainers(ctx)

	for _, container := range containersList {
		if tag == "all" {
			err := dockercontainer.StopContainer(ctx, container.Names[0])
                        stoppedContainer = append(stoppedContainer, container.Names[0])
                        if err != nil {
                                logging.Error.Println(w,"Error shown is : " , err)
                        }
		} else {
			matchValue, _ := regexp.MatchString(tag, container.Names[0])
			if matchValue == true {
				fmt.Fprintf(w, "Stopping " + container.Names[0]  + "\n")
				err := dockercontainer.StopContainer(ctx, container.Names[0])
				stoppedContainer = append(stoppedContainer, container.Names[0])
				if err != nil {
					logging.Error.Println("Error shown is : ", err )
					fmt.Fprintf(w, "Stop container operation failed for " + container.Names[0] )
				}
			}
			fmt.Fprintf(w, "Stopped container list is as below\n")
			for _, containername := range stoppedContainer {
				fmt.Fprintf(w, containername + "\n")
				logging.Info.Println(containername)
			}
		}
	}

	fmt.Fprintf(w, "Finished stopping the containers.\n")
}


// Handler for delete request
func delhandler(w http.ResponseWriter, r *http.Request) {
	var killedContainer []string
	var deallocatedPort []int

	vars := mux.Vars(r)
	tag := vars["tag"]

	logging.Info.Println(tag)

	tb, _ := db.GetTestBedFromID(ctx, tag)

	for value := range tb.Container {
		containername := tag + "-" + tb.Container[value].Image
		if tb.Container[value].Image == "mongo" || tb.Container[value].Image == "redis" {
			fmt.Fprintf(w, "Deleting container " + tb.Container[value].Image + " and deallocating ports \n")
			err := dockercontainer.RemoveContainer(ctx, containername)
			killedContainer = append(killedContainer, tb.Container[value].Image)
			deallocatedPort = append(deallocatedPort, tb.Container[value].SvcPort)

			if err != nil {
				logging.Error.Println("Error shown is : ", err.Error() )
				fmt.Fprintf(w, "Error shown is : ", err.Error() )
			} else {
				for _, containername := range killedContainer {
					fmt.Fprintf(w, containername + "\n")
					logging.Info.Println(containername)
				}
				for _, p := range deallocatedPort {
					db.DeletePortFromMeta(ctx, p)
					svcport := strconv.Itoa(p)
					//fmt.Fprintf(w, svcport + "\n")
					logging.Info.Println(svcport)
				}

				_, err := db.UpdateTestBedStatus(context.TODO(), tag, "Deleted")
			        if err != nil {
					logging.Error.Println(err)
				}
			}
		}
	}
}

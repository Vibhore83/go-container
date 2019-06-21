/*
 * Infra Provisioner - dockercontainer.go
 *
 * dockercontainer.go wrapper is used to handle docker image and container related operations.
 *     Pulling Image
 *     List Images
 *     Create Container
 *     List Container
 *     Start Container
 *     Stop Container
 *     Remove Container
 *     Inspect Container
 *
 * API version: 1.0.0
 * Author - Vibhore
 */

package dockercontainer

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"webserver/logging"
)


var (
	results []string
	// Due to incompatibility with latest client, pinning client version to 1.39 using [export DOCKER_API_VERSION='1.39']
        cli, err = client.NewEnvClient()
	wg sync.WaitGroup
)


func init() {
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

//ListDockerImages function lists all the docker images on the machine
func ListDockerImages(ctx context.Context) ( err error) {
        images, err := cli.ImageList(ctx, types.ImageListOptions{})
        if err != nil {
                panic(err)
        }

        logging.Info.Println("Looking for images")

        for _, image := range images {
		results = append(results, image.RepoTags[0])
        }
	logging.Info.Println(results)
	return err
}

//ListContainers function lists all the docker containers on the machine
func ListContainers(ctx context.Context) ([]types.Container) {
        containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
        if err != nil {
		logging.Error.Println(err)
                //panic(err)
        }
	return containers
}

//PullDockerImages function is used to pull docker images concurrently. Goroutines are used.
func PullDockerImages(ctx context.Context, imageName string, wg *sync.WaitGroup) {
        logging.Info.Println( "Pulling docker image ", imageName)

	_, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	//reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
        if err != nil {
		logging.Error.Println(err)
		err.Error()
        }

	/*_, err = io.Copy(os.Stdout, reader)
	if err != nil {
		log.Fatal(err)
	}*/
	defer wg.Done()
}

//CreateDockerContainer function is used to create docker containers. Goroutines are used for concurrent container creation.
func CreateDockerContainer(ctx context.Context, image string, tag string, hostport int) (container.ContainerCreateCreatedBody, nat.Port) {
	logging.Info.Println("Inside CreateDockerContainer")
	var hostname string
	var port nat.Port
	switch {
		case image == "mongo":
			port = "27017/tcp"
		case image == "redis":
			port = "6379/tcp"
		case image == "zookeeper":
			port = "2181/tcp"
		case image == "kafka":
			port = "9092/tcp"
	}

	hostname = tag + "-" + image
	logging.Info.Println(hostname)

	hport := strconv.Itoa(hostport)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
	                Image: image,
			Cmd:   []string{"tail", "-f", "/dev/null"},
			Tty:   true,
			Hostname: hostname,
			WorkingDir: "/root/",
			ExposedPorts: nat.PortSet{
				port: struct{}{},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				port: []nat.PortBinding{
					{
						HostIP: "0.0.0.0",
						HostPort: hport,
					},
				},
			},
		}, nil, hostname)
	if err != nil {
		logging.Error.Println("Container creation failed for container ")
                panic(err)
	} else {
		logging.Info.Println("Container created successfully for container ", resp.ID)
	}
	return resp, port
}

//StartContainer function is used to start a container
func StartContainer(ctx context.Context, resp container.ContainerCreateCreatedBody) {
	logging.Info.Println("Inside Start Container")

	err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		logging.Error.Println("Container start failed for container ", resp.ID)
		//panic(err)
	} else {
		logging.Info.Println("Container start successful for container : ", resp.ID)
	}
}

//InspectContainer function is used to inspect a container
func InspectContainer(ctx context.Context, resp string) types.ContainerJSON {
	logging.Info.Println("Inspecting container")
	var inspectData types.ContainerJSON

	inspectData, err = cli.ContainerInspect(ctx, resp )
	if err != nil {
		logging.Error.Println("Inspect command failed for the container ", resp)
		//panic(err)
	}
	return inspectData
}

//StopContainer function is used to stop a container
func StopContainer(ctx context.Context, resp string) error {
	err = cli.ContainerStop(ctx, resp, nil)
	if err != nil {
                //logging.Error.Println("Stopped container call failed for ", resp)
                //panic(err)
		err.Error()
        } else {
		logging.Info.Println("Stopped container ", resp)
	}
	return err
}

//RemoveContainer function is used to remove a container
func RemoveContainer(ctx context.Context, resp string) error {
	err := cli.ContainerRemove(ctx, resp, types.ContainerRemoveOptions{})
	if err != nil {
		//logging.Info.Println("Removal of container ", resp, " failed" )
		err.Error()
	} else {
		logging.Info.Println("Removed container ", resp )
	}
	return err
}

/*
 * All utility operations are implemented in this source file
 *
 * API version: 1.0.0
 * Contact: Arun K, Vibhore J
 */

package util

import (
	"io/ioutil"
	"strconv"
	"math/rand"
	"net"
	"os"
	"webserver/logging"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
  "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var (
	dir string
)

func init() {
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

func ReadFile(path string) string {
        logging.Info.Println(path)

        content, err := ioutil.ReadFile(path)
        if err != nil {
                logging.Error.Println(err)
        }

        logging.Info.Println("File contents: ", string(content))
	return string(content)
}

func WriteFile(path string, data []byte) bool {
	logging.Info.Println("File path to write at is " + string(path))

        err := ioutil.WriteFile(path, data, 0644 )
        if err != nil{
                logging.Error.Println(err)
		return false
        }
	return true
}

func GenerateDir() string {
	basePath := "/tmp/"
	randomNumber1 := rand.Intn(len(charset))
	randomNumber2 := rand.Intn(len(charset))

	dir = basePath + "d" + string(charset[randomNumber1]) + string(charset[randomNumber2])
	logging.Info.Println(dir)

	return(dir)
}

func RemoveDir(dir string) error {
	logging.Info.Println("Removing directory ", dir)
	err := os.RemoveAll(dir)
	return err
}

func ConvertIntSlicetoStringBoolMap(slice []int) map[string]bool {
	var dict map[string]bool
        dict = make(map[string]bool)

        for _, s := range slice {
                a := strconv.Itoa(s)
                dict[a] = true
        }
	return dict
}

// GetFreePort returns on free TCP port available on system
func GetFreePort() (int, error) {
        ln, err := net.Listen("tcp", ":0")
        if err != nil {
                return 0, err
        }
        defer ln.Close()
        return ln.Addr().(*net.TCPAddr).Port, nil
}

// GetFreePorts returns a list of requested number of free ports
func GetFreePorts(count int) ([]int, error) {
        var portList []int
        for i :=0; i < count; i++ {
                ln, err := net.Listen("tcp", ":0")
                if err != nil {
                        return nil, err
                }
                defer ln.Close()
                portList = append(portList, ln.Addr().(*net.TCPAddr).Port)
        }
        return portList, nil
}

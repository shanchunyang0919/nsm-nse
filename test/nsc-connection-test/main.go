package main

import (

	"flag"
	"fmt"
	"strconv"
	"log"
	"time"

	helm "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/helm_api"
	k8s "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/kubernetes_api"
)

const (
	RELEASE_NAME = "testing"
	NAMESPACE = "default"
	CHARTPATH = "./helm_nsc"
	SERVICE_NAME = "vl3-service"
	REPLICA_COUNT = "1"
)

var (
	restartWaitTime string
)

//yaml.NewYAMLOrJSONDecoder(reader, buffSize).Decode(k8sStruct) from k8s.io/client-go/pkg/util/yaml

func main() {

	// pass in parameters
	installChart := flag.Bool("install", false, "uninstall helm chart")
	sleepTime := flag.Int("sleep", 0, "iteration period")
	restartCount := flag.Int("count", 0, "container restart count")
	restartWaitTime := flag.Int("restart", 3, restartWaitTime)
	//flag.StringVar(&restartWaitTime, "restart", "3", "restart wait time")

	flag.Parse()


	vals := helm.CreateValues(SERVICE_NAME, REPLICA_COUNT, strconv.Itoa(*restartWaitTime))
	release := helm.CreateReleaseInfo(RELEASE_NAME, CHARTPATH, NAMESPACE)
	h := helm.InitHelmClientEndpoint(vals, release)
	// Prevent install chart when there is existing chart

	if *installChart {
		fmt.Print("installing helm chart...\n")
		h.InstallChart()
	} else {
		fmt.Print("re-installing helm chart...\n")
		h.ReinstallChart()
	}

	if *restartCount != 0 && *sleepTime != 0 {
		log.Fatal("the restart count and sleep time should be mutually exclusive")
	} else if *sleepTime > 0 {
		fmt.Printf("SLEEP...\n")
		time.Sleep(time.Second * time.Duration(*sleepTime))
		h.UninstallChart()
	} else if *restartCount > 0 {
		restartCountMode(*restartCount, *restartWaitTime)
		h.UninstallChart()
	}
}
/*
func restartCountMode(threshold int, restartTime int){
	var restartCountMap map[string]int32
	c := k8s.InitClientEndpoint()
	depName := c.GetDeploymentName()
	fmt.Printf("deployment name: %s\n", depName)
	for {
		//restartCountMap = c.GetPodRestartInfos()
		for key, value := range restartCountMap{

			// iterate through all the pods to see if the pod belongs to the same deployment
			if strings.HasPrefix(key, depName){
				fmt.Printf("pod name: %s, restart count: %v\n", key, value)
				if value >= int32(threshold){
					fmt.Printf("the restart count have reached: %v\n", threshold)
					return
				}
			}
		}
		//time.Sleep(time.Second * time.Duration(10))
		time.Sleep(time.Second * time.Duration(10 )
	}
}
 */

func restartCountMode(thresold int, restartTime int){
	client := k8s.InitClientEndpoint()
	//where we retrieve deployment
	//dep := client.GetDeployment()
	dep := client.GetDeploymentByName("busybox-vl3-service")
	fmt.Printf("deployment name: %s\n", dep.Name)
	for{
		time.Sleep(time.Second * time.Duration(15))
		//delete deployment
		client.DeleteDeployment(dep)
		time.Sleep(time.Second * time.Duration(15))

		client.CreateDeployment(dep)
		time.Sleep(time.Second * time.Duration(restartTime))
	}
}







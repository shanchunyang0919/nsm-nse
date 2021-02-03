package main

import (
	"flag"
	"fmt"

	k8s "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/kubernetes_api"
	helm "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/helm_api"
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


func main(){
	// pass in parameters
	flag.StringVar(&restartWaitTime, "restart", "3", "restart wait time")
	flag.Parse()

	fmt.Println(restartWaitTime)

	// Init helm endpoint
	vals := helm.CreateValues(SERVICE_NAME, REPLICA_COUNT, restartWaitTime)
	release := helm.CreateReleaseInfo(RELEASE_NAME, CHARTPATH, NAMESPACE)
	h := helm.InitHelmClientEndpoint(vals, release)


	h.InstallChart()

	//Inspect
	c := k8s.InitClientEndpoint()
	mmm := c.GetPodRestartInfos()

	fmt.Print(mmm)


}






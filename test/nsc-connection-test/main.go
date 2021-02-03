package main

import (
	"fmt"
	helmAPI "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/helm_api"
	k8sAPI "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/kubernetes_api"

	"os"
)

const (
	//RELEASE_NAME = "testing"
	NAMESPACE = "default"
	CHARTPATH = "./nsc-busybox"
	SERVICE_NAME = "vl3-service"
	REPLICA_COUNT = "1"
)

var (
	releaseName string
	restartWaitTime string

)


func main(){
	//testing k8s api
	c := k8sAPI.InitClientEndpoint()
	mmm := c.GetPodRestartInfos()

	fmt.Print(mmm)


	//take 3 parameters

	restartWaitTime := os.Args[1]
	releaseName := os.Args[2]



	// Init helm endpoint



	vals := helmAPI.CreateValues(SERVICE_NAME, REPLICA_COUNT, restartWaitTime)
	release := helmAPI.CreateReleaseInfo(releaseName, CHARTPATH, NAMESPACE)


	fmt.Print("erererer here ")


	h := helmAPI.InitHelmClientEndpoint(vals, release)
	fmt.Print(h)
	//releaseInfo.InstallChart()


}






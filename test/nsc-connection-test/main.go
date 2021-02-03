package main

import (
	"fmt"
	helmAPI "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/helm_api"
	k8sAPI "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/kubernetes_api"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"os"
)

const (
	namespace = "default"
	chartPath = "./nsc-busybox"
	nsmServiceName = "vl3-service"
	replicaCount = "1"
)

var (
	releaseName string
	restartWaitTime string
	//chartName string
	//iterations string
)


func main(){
	//testing k8s api
	c := k8sAPI.InitClientEndpoint()
	mmm := c.GetPodRestartInfos()

	fmt.Print(mmm)





	restartWaitTime := os.Args[1]
	releaseName := os.Args[2]

	vals := createValues(restartWaitTime)

	fmt.Print("erererer here ")

	chart, err := loader.Load(chartPath)
	if err != nil{
		os.Exit(1)
	}


	releaseInfo := createReleaseInfo(*vals, chart, releaseName)
	releaseInfo.InstallChart()


}




type ReleaseInfo struct {
	ReleaseName string
	ChartPath string
	ChartName string
	Namespace string
	Values map[string]interface{}
}




func createReleaseInfo(vals map[string]interface{}, chart *chart.Chart, relName string) *helmAPI.ReleaseInfo{
	return &helmAPI.ReleaseInfo{
		ReleaseName: relName,
		ChartPath: chartPath,
		//ChartName: chartName,
		Chart: chart,
		Namespace: namespace,
		Values: vals,
	}
}


func createValues(restartWaitTime string) *map[string]interface{}{
	return &map[string]interface{}{
		"nsm": map[string]interface{}{
			"serviceName": nsmServiceName,
		},
		"restartWaitTime": restartWaitTime,
		"replicaCount": replicaCount,
	}
}
package test_helper

import (
	"bytes"
	"context"
	"io"
	"k8s.io/client-go/rest"
	"log"
	//corev1 "k8s.io/api/core/v1"
	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const(
	vl3Namespace = "wcm-system"
	vl3NSELabel = "networkservicemesh.io/app=vl3-nse-vl3-service"
	//NSCLabel = "app=busybox-vl3-service"

)

func Help(){
	kClient := kubeapi.InitClientEndpoint(vl3Namespace)

	//temp
	//vl3podname := "vl3-nse-vl3-service-78cc5c5d9c-hvs9g"

	//req := kClient.GetPodLogsRequest(vl3podname)
	//GetLogs(req)

	AssertNSELogs(kClient, "hey" )

}

// Reads pod logs request and print out the logs with I/O package
func GetLogs(req *rest.Request){
	podLogs, err := req.Stream(context.TODO())
	if err != nil{
		log.Fatal(err)
	}
	defer podLogs.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil{
		log.Fatal(err)
	}
	logs := buf.String()
	log.Print(logs)
	//TODO: assert error message here
}

func AssertNSELogs(kClient *kubeapi.KubernetesClientEndpoint, assertMessage string){
	var req *rest.Request
	// only get list of NSE pods under wcm-system namespace
	podList := kClient.GetPodList(vl3NSELabel)

	//test
	//podList := kClient.GetPodList("test")

	log.Printf("asseting message: %v", assertMessage)

	for _, pod := range podList.Items{
		log.Println("-----pod name----",pod.Name)
		//req = kClient.GetPodLogsSinceSeconds(pod.Name, 3600)
		req = kClient.GetPodLogsTails(pod.Name, 20)

		GetLogs(req)
	}

}
















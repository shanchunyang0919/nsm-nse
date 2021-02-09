package test_helper

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"regexp"

	//corev1 "k8s.io/api/core/v1"
	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const(
	vl3Namespace = "wcm-system"
	vl3NSELabel = "networkservicemesh.io/app=vl3-nse-vl3-service"
	//NSCLabel = "app=busybox-vl3-service"
)

type Container struct{
	ContainerName string
	PodName string
	Namespace string
}

var (
	linuxPingRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
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

// get into the pod and
func ExecIntoPod(cmd []string, containerName string, podName string, namespace string, stdin io.Reader) (string, string, error){
	kconfig := kubeapi.GetKubeConfig()
	config := kubeapi.GetClientConfig(kconfig)
	kClient, err := kubernetes.NewForConfig(config)
	if err != nil{
		log.Fatal(err)
	}

	req := kClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil{
		return "", "", fmt.Errorf("error adding to scheme")
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Command:   cmd,
		Container: containerName,
		Stdin:     stdin != nil,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", fmt.Errorf("error in Stream: %v", err)
	}
	return stdout.String(), stderr.String(), nil
}

func (c *Container) Ping(destIP string, packetTransmit int) bool{
	pingCmd := "ping -c " + strconv.Itoa(packetTransmit) + " " + destIP
	cmd := []string{"sh", "-c", pingCmd}

	// assuming stdin for the command to be nil
	stdout, stderr, err := ExecIntoPod(cmd, c.ContainerName, c.PodName, c.Namespace, nil)
	if len(stderr) != 0 {
		log.Fatal("stderr:", stderr)
	}
	if err != nil{
		log.Fatal(err)
	}

	log.Print("ping logs:\n",stdout)

	matches := linuxPingRegexp.FindString(stdout)
	if matches == ""{
		// cannot transmit packet
		return false
	}
	return true
}














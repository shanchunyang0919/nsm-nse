package test_helper

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"io"
	"strings"
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
	// match ping statistics - X packets transmitted, X packets received, X% packet loss
	linuxPingRegexp = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")
)


// Reads pod logs request and print out the logs with I/O package
func GetLogs(req *rest.Request) string{
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

	return logs
}

// takes a pod name and returns the logs corresponding to the tail numbers, and also return a boolean to
// validate if the logs match the regex
func GetNSELogs(kClient *kubeapi.KubernetesClientEndpoint, assertMessage string, podName string, tails int) string{
	var req *rest.Request

	req = kClient.GetPodLogsTails(podName, tails)
	//req = kClient.GetPodLogsSinceSeconds(podName, 3600)
	logs := GetLogs(req)
	return logs
}

func AssertNotMatch(logs string, assertMessage string) bool{
	assertMsg := regexp.MustCompile(assertMessage)
	matches := assertMsg.FindString(logs)
	if matches == ""{
		return true
	}
	return false
}

func AssertMatch(logs string, assertMessage string) bool{
	assertMsg := regexp.MustCompile(assertMessage)
	matches := assertMsg.FindString(logs)
	if matches == ""{
		return false
	}
	return true
}


// Access into specific container inside a pod and execute commands
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


// Perform linux Ping command with routing compartment identifier. This method returns stream output and
// boolean which determines if the output matches regex.
func (c *Container) Ping(destIP string, packetTransmit int) (string, bool){
	pingCmd := "ping -c " + strconv.Itoa(packetTransmit) + " " + destIP
	cmd := []string{"sh", "-c", pingCmd}

	// assuming stdin for the command to be nil
	stdout, stderr, err := ExecIntoPod(cmd, c.ContainerName, c.PodName, c.Namespace, nil)
	if len(stderr) != 0 {
		log.Fatal("stderr:", stderr)
	}
	if err != nil{
		// having problems pinging
		log.Print(err)
		return "", false
	}

	matches := linuxPingRegexp.FindString(stdout)
	if matches == ""{
		// cannot transmit packet
		return "", false
	}
	// check for packet loss
	args := strings.Split(strings.Split(matches, ",")[2], "%")
	packetLossPercentage, err := strconv.Atoi(strings.TrimPrefix(args[0], " "))
	if err != nil{
		log.Fatal(err)
	}

	if packetLossPercentage != 0{
		log.Print("packet loss is not 0%")
		return "", false
	}
	return stdout, true
}














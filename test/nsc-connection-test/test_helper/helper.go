package test_helper

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
	corev1 "k8s.io/api/core/v1"
)

// A container inside a specific pod.
// In our case, it will the be container "busybox" inside pod "busybox-vl3-service" under default namespace.
type Container struct {
	ContainerName string
	PodName       string
	Namespace     string
}

var (
	// This regex match ping statistics - X packets transmitted, X packets received, X% packet loss
	PingRegex = regexp.MustCompile("\n([0-9]+) packets transmitted, ([0-9]+) packets received, ([0-9]+)% packet loss")

	packetLossTolerance = 50
	)

// Reads pod logs request and print out the logs with I/O package
func GetLogs(req *rest.Request) (string, error) {
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", errors.Wrap(err, "error steaming")
	}

	// defer podLogs.Close()
	defer func() {
		cerr := podLogs.Close()
		if cerr != nil {
			err = cerr
		}
	}()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", errors.Wrap(err, "error copying")
	}
	logs := buf.String()

	return logs, err
}

// Access into specific container inside a pod and execute commands.
// It returns stdout, stderr, and error.
func ExecIntoPod(cmd []string, containerName string, podName string, namespace string, stdin io.Reader) (string, string, error) {
	kconfig := kubeapi.GetKubeConfig()
	config := kubeapi.GetClientConfig(kconfig)

	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}

	req := kClient.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(namespace).SubResource("exec")

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return "", "", errors.New("error adding to scheme")
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
		return "", "", err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "error initiates the transport of the standard shell streams")
	}

	return stdout.String(), stderr.String(), nil
}


func (c *Container) GetNSEInterfaceIP() (string, error){
	getIPCmd := "ip a show dev nsm0"
	cmd := []string{"sh", "-c", getIPCmd}

	m := regexp.MustCompile(`\b(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

	exec, stderr, err := ExecIntoPod(cmd, c.ContainerName, c.PodName, c.Namespace, nil)
	if len(stderr) != 0 {
		return "", errors.New("std getting nsm ip address")
	}
	if err != nil {
		return "", err
	}

	nsmIP := strings.Split(m.FindString(exec), ".")

	lastByte, err := strconv.Atoi(nsmIP[3])
	if err != nil{
		return "", err
	}

	if lastByte == 255{
		return "", errors.New("last byte ip address is 255")
	}else{
		lastByte++
	}

	vl3IPSlice:= append(nsmIP[0:3], strconv.Itoa(lastByte))
	vl3IP := strings.Join(vl3IPSlice, ".")

	return vl3IP, nil
}

// Perform linux Ping command to destination IP address. This method returns stream output and
// boolean which determines if the output matches regex.
func (c *Container) Ping(destIP string, packetTransmit int) (string, bool, error) {
	pingCmd := "ping -c " + strconv.Itoa(packetTransmit) + " " + destIP
	cmd := []string{"sh", "-c", pingCmd}

	// assuming stdin for the command to be nil
	stdout, stderr, err := ExecIntoPod(cmd, c.ContainerName, c.PodName, c.Namespace, nil)
	if len(stderr) != 0 {
		return "", false, errors.New("stderr is not nil")
	}
	if err != nil {
		// having problems pinging
		return "", false, err
	}

	matches := PingRegex.FindString(stdout)
	if matches == "" {
		// cannot transmit packet
		return "", false, nil
	}
	// check for packet loss
	args := strings.Split(strings.Split(matches, ",")[2], "%")
	packetLossPercentage, errAtoi := strconv.Atoi(strings.TrimPrefix(args[0], " "))
	if errAtoi != nil {
		return "", false, errAtoi
	}

	if packetLossPercentage > packetLossTolerance {
		logrus.Printf("packet loss is greater than %v %\n", packetLossTolerance)
		return "", false, nil
	}

	return stdout, true, nil
}


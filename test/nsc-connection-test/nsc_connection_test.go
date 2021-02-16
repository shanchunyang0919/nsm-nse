package connection_test

import (
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
	//"context"

	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test"
	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_helper"

	cgo "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// Initialize mock deployment
	initPodRestartRate = 3000
	initReplicaCount   = 1

	// labels
	nsmgrLabel = "app=nsmgr-daemonset"
	nscLabel   = "app=busybox-vl3-service"
	vl3Label   = "networkservicemesh.io/app=vl3-nse-vl3-service"

	// namespaces
	nscNamespace = "default"
	wcmNamespace = "wcm-system"
	nsmNamespace = "nsm-system"

	// containername
	nscContainerName = "busybox"
	vl3ContainerName = "vl3-nse"
	nsmgrContainerName = "nsmd"

	// connectivity test
	packetTransmit = 1
)

var (
	// environment variables
	INIT_MODE string
	TIMEOUT   int
	NSE_LOG   int
	NSMGR_LOG int
	PING_LOG  string
	LOG       string
)

func setEnvironmentVariables() error {
	INIT_MODE = os.Getenv("INIT")
	log.Println("INIT MODE:", INIT_MODE)

	var err error

	TIMEOUT, err = strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil || TIMEOUT < 0 {
		return errors.New("error setting TIMEOUT")
	}
	log.Println("TIMEOUT:", TIMEOUT)

	NSE_LOG, err = strconv.Atoi(os.Getenv("NSE_LOG"))
	if err != nil || NSE_LOG < 0 {
		return errors.New("error setting NSE_LOG")
	}
	log.Println("NSE_LOG:", NSE_LOG)

	NSMGR_LOG, err = strconv.Atoi(os.Getenv("NSMGR_LOG"))
	if err != nil || NSMGR_LOG < 0 {
		return errors.New("error setting NSMGR_LOG")
	}
	log.Println("NSMGR_LOG:", NSMGR_LOG)

	PING_LOG = os.Getenv("INIT")
	log.Println("PING_LOG:", PING_LOG)

	// turn off log mode
	LOG = os.Getenv("LOG")
	if LOG == "off" {
		NSE_LOG = 0
		NSMGR_LOG = 0
		PING_LOG = "off"
	}

	return nil
}

func TestMain(m *testing.M) {
	err := setEnvironmentVariables()
	if err != nil {
		log.Fatal(err)
	}
	timeout := time.After(time.Second * time.Duration(TIMEOUT))
	done := make(chan bool)

	// it waits either the done channel to finish or timeout
	go func() {
		m.Run()
		done <- true
	}()

	select {
	case <-timeout:
		log.Fatal("nsc connection test didn't finish in time")
	case <-done:
	}
}


type BounceParameters struct {
	podRestartRate         int
	podRestartFrequency    int
	restartIterationPeriod int
	replicaCount           int
}

// NS connectivity test after the iteration of repeated bring up/down runs.
func TestConnectivity(t *testing.T) {
	if (INIT_MODE) == "on" {
		err := Init(initPodRestartRate, initReplicaCount)
		if err != nil{
			log.Fatal(err)
		}
	}
	//setup
	defaultClientEndpoint := cgo.InitClientEndpoint(metav1.NamespaceDefault)
	wcmClientEndpoint := cgo.InitClientEndpoint(wcmNamespace)
	vl3PodList := wcmClientEndpoint.GetPodListByLabel(vl3Label)

	// get list of nsmgr
	var nsmgrPodList *corev1.PodList
	var nsmClientEndpoint *cgo.KubernetesClientEndpoint
	if NSMGR_LOG > 0 {
		nsmClientEndpoint = cgo.InitClientEndpoint(nsmNamespace)
		nsmgrPodList = nsmClientEndpoint.GetPodListByLabel(nsmgrLabel)
	}

	log.Print("bouncing...")

	params := []BounceParameters{
		{
			podRestartRate:         20,
			podRestartFrequency:    0,
			restartIterationPeriod: 0,
			replicaCount:           1,
		},
		//{
		//	podRestartRate:         40,
		//	podRestartFrequency:    0,
		//	restartIterationPeriod: 20,
		//	replicaCount:           1,
		//},
		//{
		//	podRestartRate:         40,
		//	podRestartFrequency:    10,
		//	restartIterationPeriod: 0,
		//	replicaCount:           1,
		//},
	}

	for _, param := range params {
		nscDeployment, err := ReSetup(param.podRestartRate, param.replicaCount)
		log.Printf("pod restart rate: %v, pod restart frequency: %v, restart iteration period: %v," +
			"replica count: %v\n", param.podRestartRate, param.podRestartFrequency,
			param.restartIterationPeriod, param.replicaCount)
		if err != nil{
			log.Fatal(err)
		}
		bounce(nscDeployment, defaultClientEndpoint, param.podRestartRate,
			param.podRestartFrequency, param.restartIterationPeriod)
		// loggers
		if NSE_LOG > 0 {
			nsePodList := wcmClientEndpoint.GetPodListByLabel(vl3Label)
			for _, nsePod := range nsePodList.Items {
				displayPodLogs(wcmClientEndpoint, nsePod, NSE_LOG, vl3ContainerName)
			}
		}
		if nsmgrPodList != nil {
			for _, nsmgrPod := range nsmgrPodList.Items {
				displayPodLogs(nsmClientEndpoint, nsmgrPod, NSMGR_LOG, nsmgrContainerName)
			}
		}
	}

	log.Print("----- Connectivity Tests -----")

	var c *Container
	var vl3DestIP string
	var successfulConnection bool

	depForConnTest := struct {
		podRestartRate int
		replicaCount   int
	}{5000, 1}
	ReSetup(depForConnTest.podRestartRate, depForConnTest.replicaCount)

	nscPodList := defaultClientEndpoint.GetPodListByLabel(nscLabel)

	// iterate through every NSC containers to ping all NSEs
	for _, nscPod := range nscPodList.Items {
		successfulConnection = false
		c = &Container{
			ContainerName: nscContainerName,
			PodName:       nscPod.Name,
			Namespace:     nscNamespace,
		}
		for podNum, vl3Pod := range vl3PodList.Items {
			vl3DestIP = wcmClientEndpoint.GetPodIP(vl3Pod.Name)
			logs, success := c.Ping(vl3DestIP, packetTransmit)

			if PING_LOG == "on" {
				log.Printf("Pod: %v, %v\n", podNum+1, c.PodName)
				log.Printf("Ping from container \"%s\" to address %s\n",
					c.ContainerName, vl3DestIP)
				log.Println(logs)
			}
			if success {
				// at least one vl3 NSE is connected to NSC
				successfulConnection = true
				break
			}
		}
		if !successfulConnection {
			t.Fatalf("error: pod %v has no successful connections\n", nscPod.Name)
		}
	}

}

// The method contains the logic creating continuously restarting client pods
// podRestartRate: restart rate (or wait time between restarts)
// podRestartFrequency: restart iteration count
// restartIterationPeriod: restart iteration time period (mutually exclusive from iteration count)
func bounce(dep *appsv1.Deployment, endpoint *cgo.KubernetesClientEndpoint, podRestartRate int, podRestartFrequency int, restartIterationPeriod int) {
	if podRestartFrequency != 0 && restartIterationPeriod != 0 {
		endpoint.DeleteDeployment(dep)
		log.Fatal("the iteration period and pod restart count should be mutually exclusive")
	} else if restartIterationPeriod > 0 {
		log.Printf("iterating for %v seconds...", restartIterationPeriod)
		time.Sleep(time.Second * time.Duration(restartIterationPeriod))
		endpoint.ReCreateNSCDeployment(dep)
	} else if podRestartFrequency > 0 {
		restartCountMode(dep, endpoint, podRestartRate, podRestartFrequency)
	}
}

func restartCountMode( dep *appsv1.Deployment, endpoint *cgo.KubernetesClientEndpoint, podRestartRate int, podRestartFrequency int) {
	for i := 1; i <= podRestartFrequency; i++ {
		log.Printf("restart count %v...", i)
		endpoint.ReCreateNSCDeployment(dep)
		time.Sleep(time.Second * time.Duration(podRestartRate))
	}
}

// Display pod logs and container counts
func displayPodLogs(kC *cgo.KubernetesClientEndpoint, pod corev1.Pod, tails int, containerName string) {
	log.Println("display logs for pod", pod.Name)
	var req *rest.Request
	req = kC.GetPodLogsTails(pod.Name, tails, containerName)
	logs := GetLogs(req)

	log.Print(logs)
	DisplayContainersRestartCount(pod)
}

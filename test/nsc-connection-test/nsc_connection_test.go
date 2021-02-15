package connection_test

import (
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
	//"context"

	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test"
	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_helper"

	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const (

	// Initialize mock deployment
	initPodRestartTime    = 3000
	initPodRestartFreq    = 0
	initRestartIterPeriod = 0
	initReplicaCount      = 1
	defaultImageName      = "busybox:1.28"
	vl3NSELabel           = "networkservicemesh.io/app=vl3-nse-vl3-service"

	// Connectivity Test
	nscNamespace     = "default"
	nscContainerName = "busybox"
	vl3Namespace     = "wcm-system"
	nscLabel         = "app=busybox-vl3-service"
	packetTransmit   = 1
)

var (
	errMsgs   = []string{"too many open files"}
	INIT_MODE = os.Getenv("INIT")
	LOG_MODE  = os.Getenv("LOG")
	TIMEOUT   = os.Getenv("TIMEOUT")
)

func TestMain(m *testing.M) {

	log.Println("INIT MODE:", INIT_MODE)
	log.Println("LOG MODE:", LOG_MODE)
	log.Println("TIMEOUT:", TIMEOUT)
	limit, err := strconv.Atoi(TIMEOUT)
	if err != nil {
		// timeout must be set
		log.Fatal(err)
	}
	if (INIT_MODE) == "on" {
		InitSetup(initPodRestartTime, initPodRestartFreq, initRestartIterPeriod, initReplicaCount)
	}
	timeout := time.After(time.Second * time.Duration(limit))
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

type MockNSC struct {
	podRestartTime    int
	podRestartFreq    int
	restartIterPeriod int
	replicaCount      int
}

func TestLogs(t *testing.T) {
	if testing.Short() {
		t.Skip("skip log test")
	}

	vl3Client := kubeapi.InitClientEndpoint(vl3Namespace)
	NSCClient := kubeapi.InitClientEndpoint(nscNamespace)
	vl3List := vl3Client.GetPodListByLabel(vl3NSELabel)

	log.Printf("----- Logs Tests -----")
	testCases := []MockNSC{
		{
			podRestartTime:    20,
			podRestartFreq:    0,
			restartIterPeriod: 5,
			replicaCount:      1,
		},
		{
			podRestartTime:    20,
			podRestartFreq:    0,
			restartIterPeriod: 10,
			replicaCount:      1,
		},
		{
			podRestartTime:    20,
			podRestartFreq:    2,
			restartIterPeriod: 0,
			replicaCount:      1,
		},

	}
	// taking lines of logs
	var tails = 20

	// asserting this message
	// TODO: make sure what error msg it is
	var errMsg = "level=error"

	for testNum, test := range testCases {
		ReSetup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod, test.replicaCount)

		// Prints out NSC restart counts
		for _, nscPod := range NSCClient.GetPodListByLabel(nscLabel).Items {
			GetContainersRestartCount(&nscPod)
		}

		// iterate through all the NSEs to search for errors logs
		for _, pod := range vl3List.Items {
			logsCaptured := GetNSELogs(vl3Client, pod.Name, tails)

			if LOG_MODE == "on" {
				log.Printf("Test Case: %v\n", testNum)
				log.Printf("Pod restart time: %v(s), Pod restart frequency: %v(s)"+
					", restart iteration period: %v(s)\n",
					test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)

				log.Printf("Asserting message: %v", errMsg)
				log.Print(logsCaptured)
			}
			fail := AssertMatch(logsCaptured, errMsg)
			if fail {
				log.Printf("Test Case: %v FAIL\n", testNum)
				log.Fatalf("Error message \"%v\" found.", errMsg)
			}
		}
	}
}

/*
// deploy & reploy serveral amounts of pods to see if nsmgr crashes
func TestLoad(t *testing.T){
	if testing.Short(){
		t.Skip("skip load test")
	}
	log.Printf("------------ Loading Tests ------------")
	testCases := []struct{
		podRestartTime int
		podRestartFreq int
		restartIterPeriod int
		replicaCount int
	}{
		{
			podRestartTime:    20,
			podRestartFreq:    0,
			restartIterPeriod: 5,
			replicaCount:      30,
		},
	}
	for testNum, test  := range testCases{
		if LOG_MODE == "on" {
			log.Printf("Test Case: %v\n", testNum)
			ReSetup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)
			// TODO: Assert statement here
		}

	}
}

*/

// NS connectivity test after the iteration of repeated bring up/down runs.
func TestConnectivity(t *testing.T) {
	//setup
	clientWCM := kubeapi.InitClientEndpoint(vl3Namespace)
	vl3List := clientWCM.GetPodListByLabel(vl3NSELabel)

	log.Print("----- Connectivity Tests -----")

	testCases := []MockNSC{
		{
			podRestartTime:    5000,
			podRestartFreq:    0,
			restartIterPeriod: 0,
			replicaCount:      1,
		},
		{
			podRestartTime:    5000,
			podRestartFreq:    0,
			restartIterPeriod: 0,
			replicaCount:      2,
		},
	}
	for _, test := range testCases {
		var c *Container
		var vl3DestIP string
		var successfulConnection bool
		ReSetup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod, test.replicaCount)
		nscList := kubeapi.InitClientEndpoint(nscNamespace).GetPodListByLabel(nscLabel)
		// iterate through every NSC containers to ping all NSEs

		for _, pod := range nscList.Items {
			successfulConnection = false
			c = &Container{
				ContainerName: nscContainerName,
				PodName:       pod.Name,
				Namespace:     nscNamespace,
			}
			for podNum, vl3pod := range vl3List.Items {
				vl3DestIP = clientWCM.GetPodIP(vl3pod.Name)
				logs, success := c.Ping(vl3DestIP, packetTransmit)

				if LOG_MODE == "on" {
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
				t.Fatalf("error: pod %v has no successful connections\n", pod.Name)
			}
		}
	}
}



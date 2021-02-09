package connection_test

import (
	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
	"os"
	//kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
	helper "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_helper"
	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test"
	//v1 "k8s.io/api/core/v1"
	//"github.com/cisco-app-networking/wcm-common/pkg/utils/kube"
	"log"
	"testing"
)

const(

	// initialize mock deployment
	initPodRestartTime = 3000
	initPodRestartFreq = 0
	initRestartIterPeriod = 0
	initReplicaCount = 1
	defaultImageName = "busybox:1.28"
	vl3NSELabel = "networkservicemesh.io/app=vl3-nse-vl3-service"

	// Connectivity
	nscNamespace = "default"
	nscContainerName = "busybox"
	vl3Namespace = "wcm-system"
	mockNscLabels = "app=busybox-vl3-service"
	packetTransmit = 5

)

var (
	errMsgs = []string{"too many open files"}

)

func TestMain(m *testing.M) {
	log.Print("------------ NSC Connection Test ------------")

	if (os.Getenv("INIT")) == "on" {
		InitSetup(initPodRestartTime, initPodRestartFreq, initRestartIterPeriod, initReplicaCount)
	}


	m.Run()

	log.Print("------------ NSC Connection Test Ends ------------")
}

type TestCase struct{
	podRestartTime int
	podRestartFreq int
	restartIterPeriod int
	replicaCount int
}


func TestLogs(t *testing.T){

	if testing.Short(){
		t.Skip("skip test in short mode")
	}
	log.Printf("------------ General Error Message Tests ------------")
	testCases := []TestCase{
		{
			podRestartTime: 20,
			podRestartFreq: 0,
			restartIterPeriod: 5,
			replicaCount: 1,
		},
		{
			podRestartTime: 20,
			podRestartFreq: 0,
			restartIterPeriod: 10,
			replicaCount: 1,
		},
		{
			podRestartTime: 20,
			podRestartFreq: 2,
			restartIterPeriod: 0,
			replicaCount: 1,
		},
	}

	// TODO: assert error mgs
	//errMsgGeneral := "connecting failed (attempt 1/3): dial unix /run/vpp/api.sock: connect: resource temporarily unavailable"
	// errms := "level=error-"

	for testNum, test  := range testCases{
		log.Printf("\n------------ Test Case %v ------------", testNum)
		log.Printf("pod restart time: %v(s), pod restart frequency: %v(s), restart iteration period: %v(s)\n",
			test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)

		Setup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod, test.replicaCount)
		// TODO: Assert statement here
		helper.Help()

	}
/*
	log.Printf("------------ High Frequency Tests ------------")

	// TODO: High freq test
	// nsmgr => "too many open files
	////CMD: type lsof in nsmgr"
	//errMsgHighFreq := "Rejecting large frequency change of"

	testCaseHighFreq := []TestCase{
		{
			podRestartTime: 2,
			podRestartFreq: 20,
			restartIterPeriod: 0,
		},
	}
	for testNum, test  := range testCaseHighFreq{
		log.Printf("------------ Test Case %v ------------", testNum)
		log.Printf("pod restart time: %v(s), pod restart frequency: %v(s), restart iteration period: %v(s)\n",
			test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)

		Setup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)
		// TODO: Assert statement here
		helper.Help()

	}
*/
}

// NS connectivity test after the iteration of repeated bring up/down runs.
func TestConnectivity(t *testing.T){
	//setup
	nscList := kubeapi.InitClientEndpoint(nscNamespace).GetPodList(mockNscLabels)
	clientWCM := kubeapi.InitClientEndpoint(vl3Namespace)
	vl3List := clientWCM.GetPodList(vl3NSELabel)

	// deploys a long live pod
	var podRestartTime = 5000
	Setup(podRestartTime, 0, 0, 1)

	log.Print("------------ Connectivity Tests ------------")
	var c *helper.Container
	var vl3DestIP string
	var connectionCount int

	// iterate through every NSC containers to ping all NSEs
	for _, pod := range nscList.Items{
		connectionCount = 0
		c = &helper.Container{
			ContainerName: nscContainerName,
			PodName: pod.Name,
			Namespace: nscNamespace,
		}
		for _, vl3pod := range vl3List.Items {
			vl3DestIP = clientWCM.GetPodIP(vl3pod.Name)
			// TODO assert here
			if c.Ping(vl3DestIP,packetTransmit){
				connectionCount++
			}
		}
		// if there is no connections at all it will fail the test
		if connectionCount == 0{
			t.Fatal("no successful connections")
		}
	}
}
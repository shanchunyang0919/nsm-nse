package main

import (
	//kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
	helper "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_helper"
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
	defaultImageName = "busybox:1.28"
	vl3NSELabel = "networkservicemesh.io/app=vl3-nse-vl3-service"
)

var (
	errMsgs = []string{"too many open files"}

)

func TestMain(m *testing.M) {
	log.Print("------------ NSC Connection Test ------------")



	// Create a NSC busybox deployment
	//InitSetup(initPodRestartTime, initPodRestartFreq, initRestartIterPeriod)

	//{"/bin/sh", "-c", "ls", "-ll", "."}
	//cmd := []string{"/bin/sh", "-c", "ping", "127.0.0.1"}
	/*
	cmd := []string{"sh", "-c","ping -c 10 google.com"}
	containerName := "busybox"
	podName := "busybox-vl3-service-7445cc4d9b-z68xj"
*/
	c := helper.Container{
		ContainerName: "busybox",
		PodName: "busybox-vl3-service-7445cc4d9b-z68xj",
		Namespace: "default",
	}

	log.Print(c.Ping("google.com", 5))

/*
	output, _, err := helper.ExecIntoPod(cmd, containerName,podName,"default", nil)
	if err != nil{
		log.Fatal(err)
	}
*/

	log.Print("------------ NSC Connection Test Ends ------------")
}

type TestCase struct{
	podRestartTime int
	podRestartFreq int
	restartIterPeriod int
}


func TestLogs(t *testing.T){
	log.Printf("------------ General Error Message Tests ------------")
	testCases := []TestCase{
		{
			podRestartTime: 20,
			podRestartFreq: 0,
			restartIterPeriod: 5,
		},
		{
			podRestartTime: 20,
			podRestartFreq: 0,
			restartIterPeriod: 10,
		},
		{
			podRestartTime: 20,
			podRestartFreq: 2,
			restartIterPeriod: 0,
		},
		/*
		{
			podRestartTime: 20,
			podRestartFreq: 5,
			restartIterPeriod: 10,
		},*/
	}

	// TODO: assert error mgs
	//errMsgGeneral := "connecting failed (attempt 1/3): dial unix /run/vpp/api.sock: connect: resource temporarily unavailable"
	// errms := "level=error-"

	for testNum, test  := range testCases{
		log.Printf("\n------------ Test Case %v ------------", testNum)
		log.Printf("pod restart time: %v(s), pod restart frequency: %v(s), restart iteration period: %v(s)\n",
			test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)

		Setup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)
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

package connection_test

import (
	"fmt"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
	//"context"

	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test"

	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
	helper "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_helper"
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
	packetTransmit = 1

)

var (
	errMsgs = []string{"too many open files"}
	INIT_MODE = os.Getenv("INIT")
	LOG_MODE = os.Getenv("LOG")
	TIMEOUT = os.Getenv("TIMEOUT")
)

func TestMain(m *testing.M) {

	log.Println("INIT MODE:", INIT_MODE)
	log.Println("LOG MODE:", LOG_MODE)
	log.Println("TIMEOUT:", TIMEOUT)
	limit, err := strconv.Atoi(TIMEOUT)
	if err != nil{
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


func TestLogs(t *testing.T){
	if testing.Short(){
		t.Skip("skip log test")
	}

	kClient := kubeapi.InitClientEndpoint(vl3Namespace)
	vl3List := kClient.GetPodList(vl3NSELabel)

	log.Printf("----- Logs Tests -----")
	testCases := []struct{
		podRestartTime int
		podRestartFreq int
		restartIterPeriod int
		replicaCount int
	}{
		{
			podRestartTime: 20,
			podRestartFreq: 0,
			restartIterPeriod: 5,
			replicaCount: 1,
		},
		/* create deployment, delete pods, delete deployments
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
		 */
	}

	// errms := "level=error-"
	var tails = 20
	var logsMatches int
	var assertMsg = "CREATE"

	for testNum, test  := range testCases{
		logsMatches = 0
		Setup(test.podRestartTime, test.podRestartFreq, test.restartIterPeriod, test.replicaCount)

		// iterate through all the NSEs to search for matching logs
		for _, pod := range vl3List.Items {

			logsCaptured := helper.GetNSELogs(kClient, assertMsg, pod.Name, tails)

			// TODO: Watch & wait till all pods are ready

			/*
			watcher, err :=api.PersistentVolumeClaims(ns).
			       Watch(listOptions)
			    if err != nil {
			      log.Fatal(err)
			    }
			    ch := watcher.ResultChan()
			 */

			if LOG_MODE == "on" {
				log.Printf("Test Case: %v\n", testNum)
				log.Printf("Pod restart time: %v(s), Pod restart frequency: %v(s)" +
					", restart iteration period: %v(s)\n",
					test.podRestartTime, test.podRestartFreq, test.restartIterPeriod)

				log.Printf("asseting message: %v", assertMsg)
				log.Print(logsCaptured)
			}
			success := helper.AssertNotMatch(logsCaptured, assertMsg)
			if success{
				log.Printf("Test Case: %v PASS\n", testNum)
				logsMatches++
				break
			}
		}
		if logsMatches == 0{
			log.Printf("Test Case: %v FAIL\n", testNum)
			log.Fatal("no log matches")
		}
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
	clientWCM := kubeapi.InitClientEndpoint(vl3Namespace)
	vl3List := clientWCM.GetPodList(vl3NSELabel)

	log.Print("----- Connectivity Tests -----")

	// deploys a long live pod
	var podRestartTime = 5000
	Setup(podRestartTime, 0, 0, 20)


	var c *helper.Container
	var vl3DestIP string
	var connectionCount int


	//time.Sleep(30 * time.Second)

	nscList := kubeapi.InitClientEndpoint(nscNamespace).GetPodList(mockNscLabels)
	log.Printf("pod count: %v \n",len(nscList.Items))

	// iterate through every NSC containers to ping all NSEs
	for _, pod := range nscList.Items{
		fmt.Println(pod.Name)

		connectionCount = 0
		c = &helper.Container{
			ContainerName: nscContainerName,
			PodName: pod.Name,
			Namespace: nscNamespace,
		}


		for _, vl3pod := range vl3List.Items {
			vl3DestIP = clientWCM.GetPodIP(vl3pod.Name)
			logs, success := c.Ping(vl3DestIP, packetTransmit)

			if LOG_MODE == "on"{
				log.Printf("Ping from pod %s: container \"%s\" to address %s\n%s",
					c.PodName, c.ContainerName, vl3DestIP, logs)
			}
			if success{
				connectionCount++
			}
		}
		if connectionCount == 0{
			t.Fatal("no successful connections")
		}
	}
}
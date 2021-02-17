package connection_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/client-go/rest"

	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test"
	. "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_helper"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cgo "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
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
	nscContainerName   = "busybox"
	vl3ContainerName   = "vl3-nse"
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
	var err error

	TIMEOUT, err = strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil || TIMEOUT < 0 {
		return errors.Wrap(err, "error setting TIMEOUT")
	}

	logrus.Println("INIT MODE:", INIT_MODE)
	logrus.Println("TIMEOUT:", TIMEOUT)
	// turn off log mode
	LOG = os.Getenv("LOG")
	if LOG == "off" {
		NSE_LOG = 0
		NSMGR_LOG = 0
		PING_LOG = "off"
		logrus.Println("LOG:", LOG)
		return nil
	}

	NSE_LOG, err = strconv.Atoi(os.Getenv("NSE_LOG"))
	if err != nil || NSE_LOG < 0 {
		return errors.Wrap(err, "error setting NSE_LOG")
	}

	NSMGR_LOG, err = strconv.Atoi(os.Getenv("NSMGR_LOG"))
	if err != nil || NSMGR_LOG < 0 {
		return errors.Wrap(err, "error setting NSMGR_LOG")
	}

	PING_LOG = os.Getenv("PING_LOG")

	logrus.Println("NSE_LOG:", NSE_LOG)
	logrus.Println("NSMGR_LOG:", NSMGR_LOG)
	logrus.Println("PING_LOG:", PING_LOG)

	return nil
}

func TestMain(m *testing.M) {

	err := setEnvironmentVariables()
	if err != nil {
		logrus.Fatal(err)
	}

	timeout := time.After(time.Second * time.Duration(TIMEOUT))
	done := make(chan bool)

	if (INIT_MODE) == "on" {
		err := Init(initPodRestartRate, initReplicaCount)
		if err != nil {
			logrus.Fatalf("error initializing: %v", err)
		}
	}
	// it waits either the done channel to finish or timeout
	go func() {
		m.Run()
		done <- true
	}()

	select {
	case <-timeout:
		logrus.Fatal("the test didn't finish in time")
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
	g := NewWithT(t)
	defaultClientEndpoint := cgo.InitClientEndpoint(metav1.NamespaceDefault)
	wcmClientEndpoint := cgo.InitClientEndpoint(wcmNamespace)

	vl3PodList, err := wcmClientEndpoint.GetPodListByLabel(vl3Label)
	if err != nil {
		t.Error(err)
	}

	// get list of nsmgr
	var nsmgrPodList *corev1.PodList
	var nsmClientEndpoint *cgo.KubernetesClientEndpoint
	if NSMGR_LOG > 0 {
		nsmClientEndpoint = cgo.InitClientEndpoint(nsmNamespace)
		nsmgrPodList, err = nsmClientEndpoint.GetPodListByLabel(nsmgrLabel)
		if err != nil {
			t.Error(err)
		}
	}

	logrus.Print("bouncing...")

	params := []BounceParameters{
		{
			podRestartRate:         20,
			podRestartFrequency:    0,
			restartIterationPeriod: 0,
			replicaCount:           1,
		},
		{
			podRestartRate:         20,
			podRestartFrequency:    0,
			restartIterationPeriod: 10,
			replicaCount:           1,
		},
		{
			podRestartRate:         5,
			podRestartFrequency:    3,
			restartIterationPeriod: 0,
			replicaCount:           1,
		},
		{
			podRestartRate:         2,
			podRestartFrequency:    0,
			restartIterationPeriod: 0,
			replicaCount:           1,
		},
	}

	for _, param := range params {
		nscDeployment, err := ReSetup(param.podRestartRate, param.replicaCount)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.Printf("pod restart rate: %v, pod restart frequency: %v, restart iteration period: %v,"+
			"replica count: %v\n", param.podRestartRate, param.podRestartFrequency,
			param.restartIterationPeriod, param.replicaCount)

		err = bounce(nscDeployment, defaultClientEndpoint, param.podRestartRate,
			param.podRestartFrequency, param.restartIterationPeriod)
		if err != nil {
			t.Errorf("error bouncing: %v", err)
		}
		// loggers
		if NSE_LOG > 0 {
			nsePodList, err := wcmClientEndpoint.GetPodListByLabel(vl3Label)
			if err != nil {
				t.Error(err)
			}
			for _, nsePod := range nsePodList.Items {
				err = displayPodLogs(wcmClientEndpoint, nsePod, NSE_LOG, vl3ContainerName)
				if err != nil {
					logrus.Warning(err)
				}
			}
		}
		if nsmgrPodList != nil {
			for _, nsmgrPod := range nsmgrPodList.Items {
				err = displayPodLogs(nsmClientEndpoint, nsmgrPod, NSMGR_LOG, nsmgrContainerName)
				if err != nil {
					logrus.Warning(err)
				}
			}
		}
	}

	logrus.Print("----- Connectivity Tests -----")

	var c *Container
	var vl3DestIP string
	var successfulConnection bool

	depForConnTest := struct {
		podRestartRate int
		replicaCount   int
	}{5000, 2}

	_, err = ReSetup(depForConnTest.podRestartRate, depForConnTest.replicaCount)
	if err != nil {
		logrus.Fatalf(err)
	}

	nscPodList, err := defaultClientEndpoint.GetPodListByLabel(nscLabel)
	if err != nil {
		t.Error(err)
	}

	// iterate through every NSC containers to ping all NSEs
	for _, nscPod := range nscPodList.Items {
		successfulConnection = false
		c = &Container{
			ContainerName: nscContainerName,
			PodName:       nscPod.Name,
			Namespace:     nscNamespace,
		}
		for podNum, vl3Pod := range vl3PodList.Items {
			vl3DestIP, err = wcmClientEndpoint.GetPodIP(vl3Pod.Name)
			if err != nil {
				t.Error(err)
			}
			logs, success, err := c.Ping(vl3DestIP, packetTransmit)
			if err != nil {
				t.Error(err)
			}
			if PING_LOG == "on" {
				logrus.Printf("Pod: %v, %v\n", podNum+1, c.PodName)
				logrus.Printf("Ping from container \"%s\" to address %s\n",
					c.ContainerName, vl3DestIP)
				logrus.Println(logs)
			}
			if success {
				// at least one vl3 NSE is connected to NSC
				successfulConnection = true
				break
			}
		}
		g.Expect(successfulConnection).Should(Equal(true),
			"pod should have successful connections.")
	}
}

// The method contains the logic creating continuously restarting client pods
// podRestartRate: restart rate (or wait time between restarts)
// podRestartFrequency: restart iteration count
// restartIterationPeriod: restart iteration time period (mutually exclusive from iteration count)
func bounce(dep *appsv1.Deployment, endpoint *cgo.KubernetesClientEndpoint, podRestartRate int,
	podRestartFrequency int, restartIterationPeriod int) error {
	if podRestartFrequency != 0 && restartIterationPeriod != 0 {
		return errors.New("iteration period and pod restart count should be mutually exclusive")
	} else if restartIterationPeriod > 0 {
		logrus.Printf("iterating for %v seconds...", restartIterationPeriod)
		time.Sleep(time.Second * time.Duration(restartIterationPeriod))
		err := endpoint.ReCreateNSCDeployment(dep)
		if err != nil {
			return err
		}
	} else if podRestartFrequency > 0 {
		err := restartCountMode(dep, endpoint, podRestartRate, podRestartFrequency)
		if err != nil {
			return err
		}
	}
	return nil
}

func restartCountMode(dep *appsv1.Deployment, endpoint *cgo.KubernetesClientEndpoint, podRestartRate int,
	podRestartFrequency int) error {
	for i := 1; i <= podRestartFrequency; i++ {
		logrus.Printf("restart count %v...", i)
		err := endpoint.ReCreateNSCDeployment(dep)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * time.Duration(podRestartRate))
	}
	return nil
}

// Display pod logs and container counts
func displayPodLogs(kC *cgo.KubernetesClientEndpoint, pod corev1.Pod, tails int, containerName string) error {
	logrus.Println("display logs for pod", pod.Name)
	var req *rest.Request
	req = kC.GetPodLogsTails(pod.Name, tails, containerName)
	logs, err := GetLogs(req)
	if err != nil {
		return errors.Wrap(err, "fail to display pod logs")
	}
	logrus.Print(logs)
	DisplayContainersRestartCount(pod)

	return nil
}

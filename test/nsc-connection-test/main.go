package main

import (
	"log"
	"flag"
	"time"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const(
	serviceName = "vl3-nse"
	imageName = "busybox:1.28"
	replicaCount = 1

)

func main(){
	// parameters
	applyDeployment := flag.Bool("apply", false, "create a new deployment")
	podRestartTime := flag.Int("re", 0, "restart rate")
	podRestartFreq := flag.Int("freq", 0, "restart iteration count")
	restartIterPeriod := flag.Int("iter", 0, "restart iteration time period")
	flag.Parse()

	dep := BusyboxDeployment(*podRestartTime)

	deploymentClient := kubeapi.InitClientEndpoint()

	if *applyDeployment{
		log.Print("create deployment...")
		deploymentClient.CreateDeployment(dep)
	}else{
		log.Print("recreate deployment...")
		deploymentClient.ReCreateDeployment(dep)
	}

	if *podRestartFreq != 0 && *restartIterPeriod != 0 {
		log.Fatal("the iteration period and pod restart countshould be mutually exclusive")
	} else if *restartIterPeriod > 0 {
		log.Printf("iterating for %v seconds...", *restartIterPeriod)
		time.Sleep(time.Second * time.Duration(*restartIterPeriod))
	} else if *podRestartFreq > 0 {
		restartCountMode(*podRestartFreq, *podRestartTime, dep, deploymentClient)
	}
}

func restartCountMode(podRestartFreq int, podRestartTime int, dep *appsv1.Deployment, endpoint *kubeapi.KubernetesClientEndpoint){
	for i := 1; i <= podRestartFreq; i++{
		log.Printf("restart count %v...", i)
		endpoint.ReCreateDeployment(dep)
		time.Sleep(time.Second * time.Duration(podRestartTime))
	}
}


// This is busybox deployment replacing nsc helloworld for testing purposing
func BusyboxDeployment(podRestartTime int) *appsv1.Deployment{
	// type conversions to fit in appsv1.Deployment
	val := int32(replicaCount)
	var restartTimePtr *int32 = &val
	podRestartTimeStr := strconv.Itoa(podRestartTime)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "busybox-" + serviceName,
			Labels: map[string]string{
				"version": "v1",
			},
			Annotations: map[string]string{
				"ns.networkservicemesh.io": serviceName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: restartTimePtr,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "busybox-" + serviceName,
					"version": "v1",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "busybox-" + serviceName,
						"version": "v1",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: "busybox",
							Image: imageName,
							Command: []string{
								"sleep",
								podRestartTimeStr,
							},
							ImagePullPolicy: apiv1.PullIfNotPresent,
						},
					},
					RestartPolicy: apiv1.RestartPolicyAlways,
				},
			},
		},
	}
}






package main

import (
	//"fmt"
	"log"
	"flag"
	"strconv"
	//"k8s.io/utils"
	apiv1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const(
	serviceName = "vl3-nse"
	imageName = "busybox:1.28"
	replicaCount = 1

)

func main(){
	// Get input parameters
	applyDeployment := flag.Bool("apply", false, "")
	podRestartTime := flag.Int("restart", 0, "")
	podRestartFreq := flag.Int("frequency", 0, "")
	restartIterPeriod := flag.Int("iter", 0, "")
	flag.Parse()

	//check
	log.Print(*applyDeployment, *podRestartTime, *podRestartFreq, *restartIterPeriod)

	if *applyDeployment{
		log.Print("apply deployment")
	}else{
		log.Print("reapply")
	}

	dep := BusyboxDeployment(*podRestartTime)
	log.Print(dep)



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






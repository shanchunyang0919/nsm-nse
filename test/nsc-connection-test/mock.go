package connection

import (
	"errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strconv"

	cgo "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const (
	servicePort     = 5000
	servicePortName = "http"
)

var (
	serviceName = "vl3-service"
	imageName   = "busybox:1.28"
)

// Create Busybox deployment and Service
func Init(podRestartRate int, replicaCount int) error{
	log.Println("initializing...")
	dep := busyboxDeployment(podRestartRate, replicaCount)
	deploymentClient := cgo.InitClientEndpoint(corev1.NamespaceDefault)
	if dep == nil{
		return errors.New("error initializing nsc deployment")
	}
	deploymentClient.CreateDeployment(dep)

	svc := busyboxService()
	if svc == nil{
		return errors.New("error initializing nsc service")
	}
	deploymentClient.CreateService(svc)


	log.Println("finished initializing...")
	return nil
}

// recreate Busybox deployment (without creating Service again)
func ReSetup(podRestartRate int, replicaCount int) (*appsv1.Deployment, error) {
	dep := busyboxDeployment(podRestartRate, replicaCount)
	if dep == nil{
		return nil, errors.New("error creating nsc deployment")
	}

	deploymentClient := cgo.InitClientEndpoint(corev1.NamespaceDefault)

	log.Println("setup...")
	deploymentClient.ReCreateNSCDeployment(dep)
	log.Println("finished setup...")
	return dep, nil
}

// This is busybox deployment replacing nsc helloworld for testing purposing
func busyboxDeployment(podRestartRate int, replicaCount int) *appsv1.Deployment {
	val := int32(replicaCount)
	var replicaCountptr = &val
	podRestartRateStr := strconv.Itoa(podRestartRate)

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
			Replicas: replicaCountptr,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "busybox-" + serviceName,
					"version": "v1",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "busybox-" + serviceName,
						"version": "v1",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "busybox",
							Image: imageName,
							Command: []string{
								"sleep",
								podRestartRateStr,
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}
}

func busyboxService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "busybox-" + serviceName,
			Labels: map[string]string{
				"app":      "busybox-" + serviceName,
				"nsm/role": "client",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: servicePortName,
					Port: servicePort,
				},
			},
			Selector: map[string]string{
				"app": "busybox-" + serviceName,
			},
		},
	}
}

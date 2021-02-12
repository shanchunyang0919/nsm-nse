package connection

import (
	"log"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeapi "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/clientgo"
)

const (
	servicePort = 5000
	servicePortName = "http"
)

var (
	serviceName  = "vl3-service"
	imageName    = "busybox:1.28"
	busyboxPodLabel = "app=busybox-vl3-service"
)

// Create Busybox deployment and Service
func InitSetup( podRestartTime int, podRestartFreq int, restartIterPeriod int, replicaCount int){
	dep := busyboxDeployment(podRestartTime, replicaCount)
	deploymentClient := kubeapi.InitClientEndpoint(corev1.NamespaceDefault)

	log.Print("create service...")
	svc := busyboxService()
	deploymentClient.CreateService(svc)

	log.Print("create deployment...")
	deploymentClient.CreateDeployment(dep)

	controller(dep, deploymentClient, podRestartTime, podRestartFreq, restartIterPeriod)
}

// recreate Busybox deployment (without creating Service again)
func ReSetup(podRestartTime int, podRestartFreq int, restartIterPeriod int, replicaCount int) {
	dep := busyboxDeployment(podRestartTime, replicaCount)
	deploymentClient := kubeapi.InitClientEndpoint(corev1.NamespaceDefault)

	log.Print("recreate deployment...")
	deploymentClient.ReCreateNSCDeployment(dep)

	controller(dep, deploymentClient, podRestartTime, podRestartFreq, restartIterPeriod)
}

// The method contains the logic creating continuously restarting client pods
// podRestartTime: restart rate (or wait time between restarts)
// podRestartFreq: restart iteration count
// restartIterPeriod: restart iteration time period (mutually exclusive from iteration count)
func controller(dep *appsv1.Deployment, deploymentClient *kubeapi.KubernetesClientEndpoint , podRestartTime int, podRestartFreq int, restartIterPeriod int){
	if podRestartFreq != 0 && restartIterPeriod != 0 {
		deploymentClient.DeleteDeployment(dep)
		log.Fatal("the iteration period and pod restart count should be mutually exclusive")
	} else if restartIterPeriod > 0 {
		log.Printf("iterating for %v seconds...", restartIterPeriod)
		time.Sleep(time.Second * time.Duration(restartIterPeriod))
		deploymentClient.ReCreateNSCDeployment(dep)
	} else if podRestartFreq > 0 {
		restartCountMode(podRestartFreq, podRestartTime, dep, deploymentClient)
	}
}

func restartCountMode(podRestartFreq int, podRestartTime int, dep *appsv1.Deployment, endpoint *kubeapi.KubernetesClientEndpoint) {
	for i := 1; i <= podRestartFreq; i++ {
		log.Printf("restart count %v...", i)
		endpoint.ReCreateNSCDeployment(dep)
		time.Sleep(time.Second * time.Duration(podRestartTime))
	}
}

// This is busybox deployment replacing nsc helloworld for testing purposing
func busyboxDeployment(podRestartTime int, replicaCount int) *appsv1.Deployment {
	val := int32(replicaCount)
	var replicaCountptr *int32 = &val
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
								podRestartTimeStr,
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

func busyboxService() *corev1.Service{
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "busybox-" + serviceName,
			Labels: map[string]string{
				"app": "busybox-" + serviceName,
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

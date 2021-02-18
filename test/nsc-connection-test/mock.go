package connection

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
func Init(podRestartRate int, replicaCount int) error {
	logrus.Println("initializing...")
	dep := busyboxDeployment(podRestartRate, replicaCount)
	deploymentClient := cgo.InitClientEndpoint(corev1.NamespaceDefault)

	derr := deploymentClient.CreateDeployment(dep)
	if derr != nil{
		logrus.Warning(derr)
	}

	svc := busyboxService()
	serr := deploymentClient.CreateService(svc)
	if serr != nil{
		logrus.Warning(serr)
	}

	if serr != nil && derr != nil{
		return errors.New("cannot create service and deployment")
	}

	logrus.Println("finished initializing...")

	return nil
}

// recreate Busybox deployment (without creating Service again)
func ReSetup(podRestartRate int, replicaCount int) (*appsv1.Deployment, error) {
	dep := busyboxDeployment(podRestartRate, replicaCount)
	deploymentClient := cgo.InitClientEndpoint(corev1.NamespaceDefault)

	logrus.Println("setup...")

	err := deploymentClient.ReCreateNSCDeployment(dep)
	if err != nil {
		return nil, err
	}

	logrus.Println("finished setup...")

	return dep, nil
}

// This is busybox deployment replacing nsc helloworld for testing purposing
func busyboxDeployment(podRestartRate int, replicaCount int) *appsv1.Deployment {
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
			Replicas: intToint32ptr(replicaCount),
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

func intToint32ptr(i int) *int32 {
	val := int32(i)

	return &val
}


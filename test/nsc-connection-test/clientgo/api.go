package clientgo

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	HOME_ENV = "HOME"

	// used for package clientcmd to create client set
	MASTER_URL = ""
)

var (
	nscLabel         = "app=busybox-vl3-service"
	nscContainerName = "busybox"

	// wait time for the kubelet to handle deletions/creations of k8s resource (millisecond)
	waitTime time.Duration = 1000

	// grace-period time for deletions of k8s resources (shorten the time of pod's TERMINATING state)
	gracePeriod int64 = 0
)

type KubernetesClientEndpoint struct {
	Kubeconfig string
	Namespace  string
	ClientSet  *kubernetes.Clientset
}

func InitClientEndpoint(namespace string) *KubernetesClientEndpoint {
	kconfig := GetKubeConfig()
	clientSet := CreateClientSet(kconfig)
	return &KubernetesClientEndpoint{
		Kubeconfig: kconfig,
		Namespace:  namespace,
		ClientSet:  clientSet,
	}
}

func CreateClientSet(kconfig string) *kubernetes.Clientset {
	config := GetClientConfig(kconfig)
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatal(err)
	}

	return clientSet
}

func GetKubeConfig() string {
	return filepath.Join(os.Getenv(HOME_ENV), ".kube", "config")
}

func GetClientConfig(kconfig string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags(MASTER_URL, kconfig)
	if err != nil {
		logrus.Fatal(err)
	}

	return config
}

func (kc *KubernetesClientEndpoint) CreateDeployment(dep *appsv1.Deployment) error {
	_, err := kc.ClientSet.AppsV1().Deployments(kc.Namespace).Create(context.TODO(),
		dep, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "error creating deployment")
	}
	return nil
}

// Deletion of Deployment with grace period set
func (kc *KubernetesClientEndpoint) DeleteDeployment(dep *appsv1.Deployment) error {
	err := kc.ClientSet.AppsV1().Deployments(kc.Namespace).Delete(context.TODO(), dep.Name,
		metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if err != nil {
		return errors.Wrap(err, "error deleting deployment")
	}
	return nil
}

// Completely clean up busybox NSC deployment and NSC pods in zero grace period mode.
// Making deletion API calls only on development-level will not immediately delete pods.
// Must also delete pods with zero grace period to reduce the wait time of pods TERMINATING status.
func (kc *KubernetesClientEndpoint) CleanUpNSCs(dep *appsv1.Deployment) error {
	err := kc.DeleteDeployment(dep)
	if err != nil {
		return err
	}
	err = kc.DeletePodByLabel(nscLabel)
	if err != nil {
		return err
	}
	return nil
}

// Clean up the NSC deployment and NSC pods with zero grace period and re-deploy the NSC deployment.
// This method will return after all busybox containers are in READY status.
// For testing purposes only, must need to set timeout option for it.
func (kc *KubernetesClientEndpoint) ReCreateNSCDeployment(dep *appsv1.Deployment) error {
	//nscContainersCount := 1
	logrus.Warning("HERE")
	err := kc.CleanUpNSCs(dep)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * waitTime)
	err = kc.CreateDeployment(dep)
	if err != nil {
		return err
	}
	// check if all the busybox containers are running
	for {
		logrus.Warning("HERE")
		var hasNotReadyContainer bool
		podList, err := kc.GetPodListByLabel(nscLabel)
		if err != nil {
			return err
		}
		// wait till the pod is created
		if len(podList.Items) == 0 {
			time.Sleep(time.Millisecond * waitTime)
			continue
		}
		for _, pod := range podList.Items {
			//if len(pod.Status.ContainerStatuses) <= nscContainersCount {
			//	// unexpected admission error
			//	hasNotReadyContainer = true
			//	break
			//}

			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name != nscContainerName {
					continue
				}
				if !containerStatus.Ready {
					hasNotReadyContainer = true
					break
				}
			}
			if hasNotReadyContainer {
				break
			}
		}
		if hasNotReadyContainer == false {
			break
		}
		hasNotReadyContainer = true
		time.Sleep(time.Millisecond * waitTime)
	}
	logrus.Print("all nsc containers are ready...")
	return nil
}

// Delete selected pods with selected label, the grace period is set to 0 here.
func (kc *KubernetesClientEndpoint) DeletePodByLabel(label string) error {
	podList, err := kc.GetPodListByLabel(label)
	if err != nil {
		return err
	}

	watch, err := kc.ClientSet.CoreV1().Pods(kc.Namespace).Watch(context.TODO(),
		metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return err
	}

	// watch a pod until it is deleted
	for _, pod := range podList.Items {
		logrus.Printf("watching %v...\n", pod.Name)

		deleted := make(chan bool)
		go func(podName string) {
			for event := range watch.ResultChan() {
				if event.Type == "DELETED" {
					deletedPod, ok := event.Object.(*corev1.Pod)
					if !ok {
						logrus.Fatal("unexpected type")
					}
					if podName == deletedPod.Name {
						deleted <- true
					}
				}
			}
		}(pod.Name)

		err = kc.ClientSet.CoreV1().Pods(kc.Namespace).Delete(context.TODO(), pod.Name,
			metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
		if err != nil {
			return errors.Wrap(err, "error deleting pod by label")
		}

		ForLoop:
			for {
				select {
				case <-deleted:
					logrus.Println(pod.Name, "is deleted")
					break ForLoop
				}
			}
	}

	return nil
}

func (kc *KubernetesClientEndpoint) GetPodListByLabel(labels string) (*corev1.PodList, error) {
	podList, err := kc.ClientSet.CoreV1().Pods(kc.Namespace).List(context.TODO(),
		metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		return nil, errors.Wrap(err, "error getting pod list by label")
	}

	return podList, nil
}

// Same behavior as: kubectl logs <pod name> --tail=<tails_num>
func (kc *KubernetesClientEndpoint) GetPodLogsTails(podName string, tails int, containerName string) *rest.Request {
	tailsInt64 := intToint64ptr(tails)
	request := kc.ClientSet.CoreV1().Pods(kc.Namespace).GetLogs(podName,
		&corev1.PodLogOptions{TailLines: tailsInt64, Container: containerName})

	return request
}

func (kc *KubernetesClientEndpoint) CreateService(service *corev1.Service) error {
	_, err := kc.ClientSet.CoreV1().Services(kc.Namespace).Create(context.TODO(),
		service, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "error creating service")
	}

	return nil
}

func intToint64ptr(i int) *int64 {
	val := int64(i)

	return &val
}



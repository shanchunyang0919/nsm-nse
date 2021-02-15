package clientgo

import (
	"os"
	"log"
	"time"
	"context"

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
	waitTimeMs time.Duration = 1000

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
		log.Fatal(err)
	}

	return clientSet
}

func GetKubeConfig() string {
	return filepath.Join(os.Getenv(HOME_ENV), ".kube", "config")
}

func GetClientConfig(kconfig string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags(MASTER_URL, kconfig)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func (kc *KubernetesClientEndpoint) CreateDeployment(dep *appsv1.Deployment) {
	_, err := kc.ClientSet.AppsV1().Deployments(kc.Namespace).Create(context.TODO(),
		dep, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

// Deletion of Deployment with grace period set
func (kc *KubernetesClientEndpoint) DeleteDeployment(dep *appsv1.Deployment) {
	err := kc.ClientSet.AppsV1().Deployments(kc.Namespace).Delete(context.TODO(), dep.Name,
		metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if err != nil {
		log.Fatal(err)
	}
}

// Completely clean up busybox NSC deployment and NSC pods in zero grace period mode.
// Making deletion API calls only on development-level will not immediately delete pods.
// Must also delete pods with zero grace period to reduce the wait time of pods TERMINATING status.
func (kc *KubernetesClientEndpoint) CleanUpNSCs(dep *appsv1.Deployment) {
	kc.DeleteDeployment(dep)
	kc.DeletePodByLabel(nscLabel)
}

// Clean up the NSC deployment and NSC pods with zero grace period and re-deploy the NSC deployment.
// This method will return after all busybox containers are in READY status.
// For testing purposes only, must need to set timeout option for it.
func (kc *KubernetesClientEndpoint) ReCreateNSCDeployment(dep *appsv1.Deployment) {
	kc.CleanUpNSCs(dep)
	time.Sleep(time.Millisecond * waitTimeMs)
	kc.CreateDeployment(dep)
	// check if all the busybox containers are running
	for {
		var hasNotReadyContainer bool
		podList := kc.GetPodListByLabel(nscLabel)
		// wait till the pod is created
		if len(podList.Items) == 0 {
			time.Sleep(time.Millisecond * waitTimeMs)
			continue
		}
		for _, pod := range podList.Items {
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
		time.Sleep(time.Millisecond * waitTimeMs)
	}
	log.Print("all nsc containers are ready...")
}

// Delete selected pods with selected label, the grace period is set to 0 here.
func (kc *KubernetesClientEndpoint) DeletePodByLabel(label string) {
	for _, pod := range kc.GetPodListByLabel(label).Items {
		kc.ClientSet.CoreV1().Pods(kc.Namespace).Delete(context.TODO(), pod.Name,
			metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	}
}

func (kc *KubernetesClientEndpoint) GetPodListByLabel(labels string) *corev1.PodList {
	podList, err := kc.ClientSet.CoreV1().Pods(kc.Namespace).List(context.TODO(),
		metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		log.Fatal(err)
	}

	return podList
}

// Same behavior as: kubectl logs <pod name> --since=<time>
func (kc *KubernetesClientEndpoint) GetPodLogsSinceSeconds(podName string, seconds int) *rest.Request {
	secondsInt64 := intToint64ptr(seconds)
	request := kc.ClientSet.CoreV1().Pods(kc.Namespace).GetLogs(podName,
		&corev1.PodLogOptions{SinceSeconds: secondsInt64})

	return request
}

// Same behavior as: kubectl logs <pod name> --tail=<tails_num>
func (kc *KubernetesClientEndpoint) GetPodLogsTails(podName string, tails int) *rest.Request {
	tailsInt64 := intToint64ptr(tails)
	request := kc.ClientSet.CoreV1().Pods(kc.Namespace).GetLogs(podName,
		&corev1.PodLogOptions{TailLines: tailsInt64})

	return request
}

func (kc *KubernetesClientEndpoint) GetPodIP(podName string) string {
	pod, err := kc.ClientSet.CoreV1().Pods(kc.Namespace).Get(context.TODO(),
		podName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	return pod.Status.PodIP
}

func (kc *KubernetesClientEndpoint) CreateService(service *corev1.Service) {
	_, err := kc.ClientSet.CoreV1().Services(kc.Namespace).Create(context.TODO(),
		service, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

func intToint64ptr(i int) *int64 {
	val := int64(i)

	return &val
}

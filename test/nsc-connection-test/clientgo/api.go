package clientgo

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	HOME_ENV  = "HOME"

	// used for package clientcmd to create client set
	MASTER_URL = ""
)




type KubernetesClientEndpoint struct{
	Kubeconfig string
	Namespace string
	ClientSet *kubernetes.Clientset
}

func (kc *KubernetesClientEndpoint) CreateDeployment(dep *appsv1.Deployment){
	createDeployment(kc.ClientSet, kc.Namespace, dep)
}

func (kc *KubernetesClientEndpoint) DeleteDeployment(dep *appsv1.Deployment){
	deleteDeployment(kc.ClientSet, kc.Namespace, dep.Name)
}

// clean up delployments && NSC pods
func (kc *KubernetesClientEndpoint) CleanUpNSC(dep *appsv1.Deployment){
	kc.DeleteDeployment(dep)
	nscLabel := "app=busybox-vl3-service"
	kc.DeletePodByLabel(nscLabel)
}

// clean up
func (kc *KubernetesClientEndpoint) ReCreateNSCDeployment(dep *appsv1.Deployment) {
	nscLabels := "app=busybox-vl3-service"
	nscContainerName := "busybox"
	kc.CleanUpNSC(dep)
	time.Sleep(time.Second * 2)
	createDeployment(kc.ClientSet, kc.Namespace, dep)
	// check if all the busybox containers are running
	for {
		var hasNotReadyContainer bool
		podList := kc.GetPodList(nscLabels)
		// wait till the pod is created
		if len(podList.Items) == 0{
			time.Sleep(time.Millisecond * 250)
			continue
		}
		for _, pod := range podList.Items{
			for _, containerStatus := range pod.Status.ContainerStatuses{
				if containerStatus.Name != nscContainerName{
					continue
				}
				if !containerStatus.Ready{
					hasNotReadyContainer = true
					break
				}
			}
		}
		if hasNotReadyContainer == false{
			break
		}
		hasNotReadyContainer = true
		time.Sleep(time.Millisecond * 250)
	}
}


/*
func (kc *KubernetesClientEndpoint) ReCreateDeployment(dep *appsv1.Deployment){

	deleteDeployment(kc.ClientSet, kc.Namespace, dep.Name)
	for{
		podList := kc.GetPodList("app=busybox-vl3-service")
		var alivePodCount int
		for _, pod := range podList.Items{
			if pod.DeletionTimestamp == nil{
				alivePodCount++
			}
		}
		if alivePodCount == 0{
			break
		}
		alivePodCount = 0
		time.Sleep(time.Millisecond * 250)
	}
	createDeployment(kc.ClientSet, kc.Namespace, dep)
	for{
		podList := kc.GetPodList("app=busybox-vl3-service")
		var hasNotReadyContainer bool
		for _, pod := range podList.Items{
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready{
					hasNotReadyContainer = true
					break
				}
			}
		}
		if !hasNotReadyContainer{
			break
		}
		hasNotReadyContainer = true
		time.Sleep(time.Millisecond * 250)
	}
}
*/




// delete seleced pods with no grace-period (shortened terminal state)
func (kc *KubernetesClientEndpoint) DeletePodByLabel(label string) {
	var gracePeriod int64 = 0
	for _, pod := range kc.GetPodList(label).Items{
		kc.ClientSet.CoreV1().Pods(kc.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	}
}

func (kc *KubernetesClientEndpoint) GetPodList(labels string) *corev1.PodList{
	return getPodList(kc.ClientSet, kc.Namespace, labels)
}

func (kc *KubernetesClientEndpoint) GetPodLogsSinceSeconds(podName string, seconds int) *rest.Request{
	return getPodLogsSinceSecond(kc.ClientSet, kc.Namespace, podName, seconds)
}

func (kc *KubernetesClientEndpoint) GetPodLogsTails(podName string, tails int) *rest.Request{
	return getPodLogsTails(kc.ClientSet, kc.Namespace, podName, tails)
}

func (kc *KubernetesClientEndpoint) GetPodIP(podName string) string{
	return getPodIP(kc.ClientSet, kc.Namespace, podName)
}

func (kc *KubernetesClientEndpoint) GetContainerID(podName string, imageName string) string{
	return getContainerID(kc.ClientSet, kc.Namespace, podName, imageName)
}

func (kc *KubernetesClientEndpoint) CreateService(service *corev1.Service){
	_, err := kc.ClientSet.CoreV1().Services(kc.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil{
		log.Fatal(err)
	}
}


func InitClientEndpoint(namespace string) *KubernetesClientEndpoint{
	kconfig := GetKubeConfig()
	clientSet := createClientset(kconfig)
	return &KubernetesClientEndpoint{
		Kubeconfig: kconfig,
		Namespace: namespace,
		ClientSet: clientSet,
	}
}

func createClientset(kconfig string) *kubernetes.Clientset {
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

func GetClientConfig(kconfig string) *rest.Config{
	config, err := clientcmd.BuildConfigFromFlags(MASTER_URL, kconfig)
	if err != nil {
		log.Fatal(err)
	}
	return config
}



package clientgo

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (kc *KubernetesClientEndpoint) ReCreateDeployment(dep *appsv1.Deployment){
	deleteDeployment(kc.ClientSet, kc.Namespace, dep.Name)
	createDeployment(kc.ClientSet, kc.Namespace, dep)
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

func (kc *KubernetesClientEndpoint) CreateService(service *corev1.Service){
	_, err := kc.ClientSet.CoreV1().Services(kc.Namespace).Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil{
		log.Fatal(err)
	}
}

func InitClientEndpoint(namespace string) *KubernetesClientEndpoint{
	kconfig := getKubeConfig()
	clientSet := createClientset(kconfig)
	return &KubernetesClientEndpoint{
		Kubeconfig: kconfig,
		Namespace: namespace,
		ClientSet: clientSet,
	}
}

func createClientset(kconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags(MASTER_URL, kconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return clientSet
}

func getKubeConfig() string {
	return filepath.Join(os.Getenv(HOME_ENV), ".kube", "config")
}



package kubernetes_api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	//v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	//hey "github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/kubernetes_api"
)

const (
	HOME_ENV  = "HOME"
	NAMESPACE = "default"

	// used for package clientcmd
	MASTER_URL = ""
)

type Utils interface{
	GetPods() *corev1.PodList
	DisplayPods()
	GetPodRestartInfos() map[string]int32

}

type KubernetesClientEndpoint struct{
	kubeconfig string
	namespace string
	clientset *kubernetes.Clientset

}


func (kc *KubernetesClientEndpoint) GetPods() (podList *corev1.PodList){
	podList = GetPodList(kc.clientset, kc.namespace)
	return
}

func (kc *KubernetesClientEndpoint) DisplayPods() {
	podList := GetPodList(kc.clientset, kc.namespace)
	PrintPodList(kc.clientset, kc.namespace, podList)
}

// returns a map using pod name and key and restart count as value
func (kc *KubernetesClientEndpoint) GetPodRestartInfos() map[string]int32{
	var restartCount int32
	m := make(map[string]int32, 0)
	podList := GetPodList(kc.clientset, kc.namespace)
	for _, pod := range podList.Items {
		restartCount = GetPodRestartCount(kc.clientset, pod.Name, kc.namespace)
		m[pod.Name] = restartCount
	}

	//test
	fmt.Print(m)
	return m
}

func createClientset(kconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags(MASTER_URL, kconfig)
	if err != nil {
		log.Fatal("cannot build config from flags")
	}

	log.Print("create clientset...")
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("cannot create clientset")
	}

	return clientSet
}

func getKubeConfig() string {
	return filepath.Join(os.Getenv(HOME_ENV), ".kube", "config")
}

//export function
func InitClientEndpoint() *KubernetesClientEndpoint{
	kconfig := getKubeConfig()
	clientSet := createClientset(kconfig)
	return &KubernetesClientEndpoint{
		kubeconfig: kconfig,
		namespace: NAMESPACE,
		clientset: clientSet,
	}
}
	// Build config from flags

	//List deployments

	/*
	depList := GetDeploymentList(clientSet, namespace)

	PrintDeploymentList(depList)

	podList := GetPodList(clientSet, namespace)
	PrintPodList(clientSet, namespace, podList)

}*/

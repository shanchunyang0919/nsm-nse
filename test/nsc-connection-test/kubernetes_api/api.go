package kubernetes_api

import (
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
	Kubeconfig string
	Namespace string
	ClientSet *kubernetes.Clientset

}


func (kc *KubernetesClientEndpoint) GetPods() (podList *corev1.PodList){
	podList = getPodList(kc.ClientSet, kc.Namespace)
	return
}

func (kc *KubernetesClientEndpoint) DisplayPods() {
	podList := getPodList(kc.ClientSet, kc.Namespace)
	printPodList(kc.ClientSet, kc.Namespace, podList)
}

// returns a map using pod name and key and restart count as value
func (kc *KubernetesClientEndpoint) GetPodRestartInfos() map[string]int32{
	var restartCount int32
	m := make(map[string]int32, 0)
	podList := getPodList(kc.ClientSet, kc.Namespace)
	for _, pod := range podList.Items {
		restartCount = getPodRestartCount(kc.ClientSet, pod.Name, kc.Namespace)
		m[pod.Name] = restartCount
	}

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

//public function
func InitClientEndpoint() *KubernetesClientEndpoint{
	kconfig := getKubeConfig()
	clientSet := createClientset(kconfig)
	return &KubernetesClientEndpoint{
		Kubeconfig: kconfig,
		Namespace: NAMESPACE,
		ClientSet: clientSet,
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

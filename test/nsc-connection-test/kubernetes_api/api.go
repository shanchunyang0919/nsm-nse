package kubernetes_api

import (
	"log"
	"os"
	"path/filepath"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//typev1 "k8s.io/client-go/kubernetes/typed/core/v1"

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


}

type KubernetesClientEndpoint struct{
	kubeconfig string
	namespace string
	clientset *kubernetes.Clientset

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

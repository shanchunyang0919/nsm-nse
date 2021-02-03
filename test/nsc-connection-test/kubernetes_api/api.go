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
	HOME_ENV = "HOME"
	namespace = "default"

	// used for package clientcmd
	masterURL = ""


)

func createClientset(kconfig string) *kubernetes.Clientset{
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kconfig)
	if err != nil{
		log.Fatal("cannot build config from flags")
	}

	log.Print("create clientset...")
	// Create clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil{
		log.Fatal("cannot create clientset")
	}
	return clientSet
}


func getKubeConfig() (kconfig string){
	kconfig = filepath.Join(os.Getenv(HOME_ENV), ".kube", "config")
	return
}

func API(){
	// Generate kubeconfig path
	kconfig := getKubeConfig()


	// Build config from flags



	//List deployments
	depList := GetDeploymentList(clientSet, namespace)

	PrintDeploymentList(depList)

	podList := GetPodList(clientSet, namespace)
	PrintPodList(podList)



}


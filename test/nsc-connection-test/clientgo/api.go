package clientgo

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/apps/v1"
)

const (
	HOME_ENV  = "HOME"
	NAMESPACE = "default"

	// used for package clientcmd
	MASTER_URL = ""
)

type KubernetesClientEndpoint struct{
	Kubeconfig string
	Namespace string
	ClientSet *kubernetes.Clientset
}

func (kc *KubernetesClientEndpoint) CreateDeployment(dep *v1.Deployment){
	createDeployment(kc.ClientSet, kc.Namespace, dep)
}

func (kc *KubernetesClientEndpoint) DeleteDeployment(dep *v1.Deployment){
	deleteDeployment(kc.ClientSet, kc.Namespace, dep.Name)
}

func (kc *KubernetesClientEndpoint) ReCreateDeployment(dep *v1.Deployment){
	deleteDeployment(kc.ClientSet, kc.Namespace, dep.Name)
	createDeployment(kc.ClientSet, kc.Namespace, dep)
}

func InitClientEndpoint() *KubernetesClientEndpoint{
	kconfig := getKubeConfig()
	clientSet := createClientset(kconfig)
	return &KubernetesClientEndpoint{
		Kubeconfig: kconfig,
		Namespace: NAMESPACE,
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
	//return "~/go/src/github.com/cisco-app-networking/nsm-nse/test/nsc-connection-test/test_env_setup/cluster1-demo"
}



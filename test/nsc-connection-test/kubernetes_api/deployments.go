package kubernetes_api

import(
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"log"
	"context"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

func GetDeploymentList(clientSet *kubernetes.Clientset, namespace string) *v1.DeploymentList{
	dep, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(),metav1.ListOptions{})
	if err != nil{
		log.Fatal("cannot list deployments")
	}
	return dep
}


// Print all deployment names
func PrintDeploymentList (depList *v1.DeploymentList) {
	fmt.Print("print deployment names...\n")
	for _, dep := range depList.Items{
		fmt.Printf(dep.Name + "\n")
	}
}
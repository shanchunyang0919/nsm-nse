package kubernetes_api

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

func GetDeploymentList(clientSet *kubernetes.Clientset, namespace string) *v1.DeploymentList {
	dep, err := clientSet.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal("cannot list deployments")
	}
	return dep
}

// Print all deployment names
func PrintDeploymentList(depList *v1.DeploymentList) {
	fmt.Print("print deployment names...\n")
	for _, dep := range depList.Items {
		fmt.Printf(dep.Name + "\n")
	}
}

/*
// Delete the deployment instead of monitoring
func DeleteDeployment(clientSet *kubernetes.Clientset, namespace string, depName string) {
	err := clientSet.AppsV1().Deployments(namespace).Delete(context.TODO(), depName, metav1.DeleteOptions{})
	if err != nil{
		log.Fatalf("cannot delete deployment %s", depName)
	}

}
*/

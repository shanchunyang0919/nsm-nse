package clientgo

import (
	"log"
	"context"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

// Delete the deployment instead of monitoring
func deleteDeployment(clientSet *kubernetes.Clientset, namespace string, depName string) {
	err := clientSet.AppsV1().Deployments(namespace).Delete(context.TODO(), depName, metav1.DeleteOptions{})
	if err != nil{
		log.Fatal(err)
	}
}

func createDeployment(clientSet *kubernetes.Clientset, namespace string, dep *v1.Deployment) {
	_, err := clientSet.AppsV1().Deployments(namespace).Create(context.TODO(), dep, metav1.CreateOptions{})
	if err != nil{
		log.Fatal(err)
	}
}


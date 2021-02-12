package clientgo

import (
	"log"
	"context"

	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)



// Delete the deployment instead of monitoring
func deleteDeployment(clientSet *kubernetes.Clientset, namespace string, depName string) {
	var gracePeriod int64 = 0
	err := clientSet.AppsV1().Deployments(namespace).Delete(context.TODO(), depName,
		metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
	if err != nil{
		log.Fatal(err)
	}
}

func createDeployment(clientSet *kubernetes.Clientset, namespace string, dep *appsv1.Deployment) {
	_, err := clientSet.AppsV1().Deployments(namespace).Create(context.TODO(), dep, metav1.CreateOptions{})
	if err != nil{
		log.Fatal(err)
	}
}


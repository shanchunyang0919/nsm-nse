package kubernetes_api

import (

	"fmt"
	//v1 "k8s.io/api/apps/v1"
	"log"
	"context"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)


func GetPodList(clientSet *kubernetes.Clientset, namespace string) *corev1.PodList{
	podList, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil{
		log.Fatal("error getting pod list")
	}
	return podList
}


func PrintPodList(podList *corev1.PodList){
	fmt.Print("print pod names,,,\n")
	for _, pod := range podList.Items {
		fmt.Printf(pod.Name + "\n")
	}
}
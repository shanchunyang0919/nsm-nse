package kubernetes_api

import (
	"fmt"
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//v1 "k8s.io/api/apps/v1"
	"log"
)


func GetPodList(clientSet *kubernetes.Clientset, namespace string) *corev1.PodList{
	podList, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil{
		log.Fatal("error getting pod list")
	}
	return podList
}


func PrintPodList(clientSet *kubernetes.Clientset, namespace string, podList *corev1.PodList){
	fmt.Print("print pod names and restart counts\n")
	var restartCount int32

	for _, pod := range podList.Items {
		restartCount = GetPodRestartCount(clientSet, pod.Name, namespace)
		fmt.Printf("%s, restart: %v\n", pod.Name, restartCount)
	}
}

func GetPodRestartCount(clientSet *kubernetes.Clientset, podName string, namespace string) int32{
	pd, err := clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil{
		log.Fatal("cannot get pod info")
	}
	return pd.Status.ContainerStatuses[0].RestartCount
}




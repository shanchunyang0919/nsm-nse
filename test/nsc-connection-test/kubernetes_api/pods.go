package kubernetes_api

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

const (
	// used for getting container restart count
	containerStatusListIdx = 0
)

func GetPodList(clientSet *kubernetes.Clientset, namespace string) *corev1.PodList {
	podList, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal("error getting pod list\n")
	}

	return podList
}

func PrintPodList(clientSet *kubernetes.Clientset, namespace string, podList *corev1.PodList) {
	var restartCount int32

	fmt.Print("print pod names and restart counts\n")
	for _, pod := range podList.Items {
		restartCount = GetPodRestartCount(clientSet, pod.Name, namespace)
		fmt.Printf("%s, restart: %v\n", pod.Name, restartCount)
	}
}

func GetPodRestartCount(clientSet *kubernetes.Clientset, podName string, namespace string) int32 {
	pd, err := clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("cannot get pod info %s\n", podName)
	}

	return pd.Status.ContainerStatuses[containerStatusListIdx].RestartCount
}

/*
// could delete pod instead of deleting deployment or uninstall helm chart
func DeletePod(clientSet *kubernetes.Clientset, podName string, namespace string){
	err := clientSet.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil{
		log.Fatalf("cannot delete pod %s\n", podName)
	}
}
*/

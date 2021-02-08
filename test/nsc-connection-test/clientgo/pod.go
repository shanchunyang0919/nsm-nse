package clientgo

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

// Get the pod list from the corresponding namespace and labels
func getPodList(clientSet *kubernetes.Clientset, namespace string, labels string) *corev1.PodList {
	podList, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labels})
	if err != nil {
		log.Fatal("err")
	}

	return podList
}


func getPodIP(clientSet *kubernetes.Clientset, namespace string, podName string) string{
	pod, err := clientSet.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil{
		log.Fatal(err)
	}

	return pod.Status.PodIP
}

// kubectl logs --since
func getPodLogsSinceSecond(clientSet *kubernetes.Clientset, namespace string, podName string, seconds int) *rest.Request{
	// Type conversion to int64 pointer
	secondsPtr := intToint64ptr(seconds)

	request := clientSet.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{SinceSeconds: secondsPtr})

	return request
}

// kubectl logs --tail
func getPodLogsTails(clientset *kubernetes.Clientset, namespace string, podName string, tails int) *rest.Request{
	// Type conversion
	tailsPtr := intToint64ptr(tails)

	request := clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{TailLines: tailsPtr})

	return request

}

// get container ID from imageName
func getContainerID(clientset *kubernetes.Clientset, namespace string, podName string, imageName string) string{

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil{
		log.Fatal(err)

	}
	for _, container := range pod.Status.ContainerStatuses{
		if container.Image == imageName{
			return container.ContainerID
		}
	}
	return ""
}

// type conversion helper method
func intToint64ptr(i int) *int64{
	val := int64(i)

	return &val
}


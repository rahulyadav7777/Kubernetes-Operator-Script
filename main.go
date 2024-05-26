package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Parse command-line flags
	kubeconfig := flag.String("kubeconfig", "/home/jaihind/.kube/config", "Path to the kubeconfig file")
	namespace := flag.String("namespace", "", "Namespace to clean up (leave empty for all namespaces)")
	labelSelector := flag.String("label-selector", "", "Label selector to filter pods (leave empty for all pods)")
	flag.Parse()

	// Initialize Kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Define cleanup intervals
	cleanupInterval := 10 * time.Second
	cleanupTicker := time.NewTicker(cleanupInterval)
	defer cleanupTicker.Stop()

	// Cleanup loop
	for {
		select {
		case <-cleanupTicker.C:
			// Perform cleanup tasks
			cleanupEvictedPods(clientset, *namespace, *labelSelector)
			cleanupCrashLoopBackOffPods(clientset, *namespace, *labelSelector)
			cleanupImagePullErrorPods(clientset, *namespace, *labelSelector)
			cleanupFailedPods(clientset, *namespace, *labelSelector)
		}
	}
}

// cleanupEvictedPods cleans up evicted pods
func cleanupEvictedPods(clientset *kubernetes.Clientset, namespace string, labelSelector string) {
	fmt.Println("Cleaning up evicted pods...")
	// List pods with the specified namespace and label selector
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing pods: %v\n", err)
		return
	}

	// Delete evicted pods
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Failed" && pod.Status.Reason == "Evicted" {
			err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting pod %s in namespace %s: %v\n", pod.Name, pod.Namespace, err)
			} else {
				fmt.Printf("Deleted evicted pod %s in namespace %s\n", pod.Name, pod.Namespace)
			}
		}
	}
}

// cleanupCrashLoopBackOffPods cleans up pods in CrashLoopBackOff state
func cleanupCrashLoopBackOffPods(clientset *kubernetes.Clientset, namespace string, labelSelector string) {
	fmt.Println("Cleaning up pods in CrashLoopBackOff state...")
	// List pods with the specified namespace and label selector
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing pods: %v\n", err)
		return
	}

	// Delete pods in CrashLoopBackOff state
	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting pod %s in namespace %s: %v\n", pod.Name, pod.Namespace, err)
				} else {
					fmt.Printf("Deleted pod %s in CrashLoopBackOff state in namespace %s\n", pod.Name, pod.Namespace)
				}
				break
			}
		}
	}
}

// cleanupImagePullErrorPods cleans up pods with ImagePullError
func cleanupImagePullErrorPods(clientset *kubernetes.Clientset, namespace string, labelSelector string) {
	fmt.Println("Cleaning up pods with ImagePullError...")
	// List pods with the specified namespace and label selector
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing pods: %v\n", err)
		return
	}

	// Delete pods with ImagePullError
	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "ImagePullBackOff" {
				err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting pod %s in namespace %s: %v\n", pod.Name, pod.Namespace, err)
				} else {
					fmt.Printf("Deleted pod %s with ImagePullError in namespace %s\n", pod.Name, pod.Namespace)
				}
				break
			}
		}
	}
}

// cleanupFailedPods cleans up pods in Failed state
func cleanupFailedPods(clientset *kubernetes.Clientset, namespace string, labelSelector string) {
	fmt.Println("Cleaning up pods in Failed state...")
	// List pods with the specified namespace and label selector
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing pods: %v\n", err)
		return
	}

	// Delete pods in Failed state
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Failed" && pod.Status.Reason != "Evicted" {
			err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting pod %s in namespace %s: %v\n", pod.Name, pod.Namespace, err)
			} else {
				fmt.Printf("Deleted pod %s in Failed state in namespace %s\n", pod.Name, pod.Namespace)
			}
		}
	}
}

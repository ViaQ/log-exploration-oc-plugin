package k8sresources

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetDaemonSetPodsList(clientset kubernetes.Interface, podList *[]string, targetDaemonset string, namespace string) error {

	var requiredDaemonSet appsv1.DaemonSet
	daemonsets, _ := clientset.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})

	requiredDaemonSetFound := false
	for _, daemonset := range daemonsets.Items {
		if daemonset.ObjectMeta.Name == targetDaemonset {
			requiredDaemonSetFound = true
			requiredDaemonSet = daemonset
			break
		}
	}

	if !requiredDaemonSetFound {
		if len(namespace) > 0 {
			return fmt.Errorf("daemon set \"%v\" not found in namespace \"%v\"", targetDaemonset, namespace)
		} else {
			return fmt.Errorf("daemon set \"%v\" not found", targetDaemonset)
		}
	}

	labelSelector := labels.Set(requiredDaemonSet.Spec.Selector.MatchLabels)

	options := metav1.ListOptions{
		LabelSelector: string(labelSelector.AsSelector().String()),
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), options)

	if err != nil {
		return fmt.Errorf("an error occurred while fetching daemon set pods: %v", err)
	}
	for _, pod := range (*pods).Items {
		*podList = append(*podList, pod.Name)
	}
	return nil
}

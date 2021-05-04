package k8sresources

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetStatefulSetPodsList(clientset kubernetes.Interface, podList *[]string, targetStatefulSet string, namespace string) error {

	var requiredStatefulSet appsv1.StatefulSet
	requiredStatefulSetFound := false
	statefulSets, _ := clientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})

	for _, statefulSet := range statefulSets.Items {
		if statefulSet.ObjectMeta.Name == targetStatefulSet {
			requiredStatefulSetFound = true
			requiredStatefulSet = statefulSet
			break
		}
	}

	if !requiredStatefulSetFound {
		if len(namespace) > 0 {
			return fmt.Errorf("stateful set \"%v\" not found in namespace \"%v\"", targetStatefulSet, namespace)
		} else {
			return fmt.Errorf("stateful set \"%v\" not found", targetStatefulSet)
		}
	}

	labelSelector := labels.Set(requiredStatefulSet.Spec.Selector.MatchLabels)

	options := metav1.ListOptions{
		LabelSelector: string(labelSelector.AsSelector().String()),
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), options)
	if err != nil {
		return fmt.Errorf("an error occurred while fetching stateful set pods: %v", err)
	}
	for _, pod := range (*pods).Items {
		*podList = append(*podList, pod.Name)
	}
	return nil
}

package k8sresources

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetDeploymentPodsList(clientset kubernetes.Interface, podList *[]string, targetDeployment string, namespace string) error {

	var requiredDeployment appsv1.Deployment
	deployments, _ := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	requiredDeploymentFound := false

	for _, deployment := range deployments.Items {
		if deployment.ObjectMeta.Name == targetDeployment {
			requiredDeploymentFound = true
			requiredDeployment = deployment
			break
		}
	}

	if !requiredDeploymentFound {
		if len(namespace) > 0 {
			return fmt.Errorf("deployment \"%v\" not found in namespace \"%v\"", targetDeployment, namespace)
		} else {
			return fmt.Errorf("deployment \"%v\" not found", targetDeployment)
		}
	}

	labelSelector := labels.Set(requiredDeployment.Spec.Selector.MatchLabels)

	options := metav1.ListOptions{
		LabelSelector: string(labelSelector.AsSelector().String()),
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), options)
	if err != nil {
		return fmt.Errorf("an error occurred while fetching deployment pods: %v", err)
	}
	for _, pod := range (*pods).Items {
		*podList = append(*podList, pod.Name)
	}
	return nil
}

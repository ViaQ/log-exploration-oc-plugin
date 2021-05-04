package k8sresources

import (
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
)

func GetResourcesPodList(kubernetesOptions *client.KubernetesOptions, resources *Resources, namespace string) ([]string, error) {

	var podList []string

	if resources.IsDeployment {
		err := GetDeploymentPodsList(kubernetesOptions.Clientset, &podList, resources.Name, namespace)
		if err != nil {
			return nil, err
		}
		return podList, nil
	}

	if resources.IsDaemonSet {
		err := GetDaemonSetPodsList(kubernetesOptions.Clientset, &podList, resources.Name, namespace)
		if err != nil {
			return nil, err
		}
		return podList, nil
	}

	if resources.IsStatefulSet {
		err := GetStatefulSetPodsList(kubernetesOptions.Clientset, &podList, resources.Name, namespace)
		if err != nil {
			return nil, err
		}
		return podList, nil
	}

	if resources.IsPod {
		podList = append(podList, resources.Name)
		return podList, nil
	}
	return nil, nil
}

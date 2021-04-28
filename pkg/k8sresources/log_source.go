package k8sresources

import (
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
)

func GetResourcesPodList(kubernetesOptions *client.KubernetesOptions, resources *Resources, namespace string) ([]string, error) {

	var podList []string

	if len(resources.Deployment) > 0 {
		err := GetDeploymentPodsList(kubernetesOptions.Clientset, &podList, resources.Deployment, namespace)
		if err != nil {
			return nil, err
		}
	}

	if len(resources.DaemonSet) > 0 {
		err := GetDaemonSetPodsList(kubernetesOptions.Clientset, &podList, resources.DaemonSet, namespace)
		if err != nil {
			return nil, err
		}
	}

	if len(resources.StatefulSet) > 0 {
		err := GetStatefulSetPodsList(kubernetesOptions.Clientset, &podList, resources.StatefulSet, namespace)
		if err != nil {
			return nil, err
		}
	}
	return podList, nil
}

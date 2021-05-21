package k8sresources

import (
	"context"
	"fmt"
	"testing"

	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetResourcesPodList(t *testing.T) {
	tests := []struct {
		TestName    string
		ShouldFail  bool
		Resources   map[string]string
		TestPodList []string
		Namespace   string
		Error       error
	}{
		{
			"Resource doesn't exist",
			false,
			map[string]string{"Deployment": "dummy-deployment"},
			[]string{},
			"openshift-logging",
			fmt.Errorf("deployment \"dummy-deployment\" not found in namespace \"openshift-logging\""),
		},
		{
			"Resources are present",
			false,
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"openshift-deployment"},
			"openshift-logging",
			nil,
		},
	}

	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "openshift-deployment",
				Namespace:   "openshift-logging",
				Annotations: map[string]string{},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"name": "logging"},
				},
			},
		})

	clientset.CoreV1().Pods("openshift-logging").Create(context.TODO(),
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name:        "openshift-deployment",
			Namespace:   "openshift-logging",
			Annotations: map[string]string{},
			Labels:      map[string]string{"name": "logging"},
		},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "logging",
					},
				},
			},
		}, metav1.CreateOptions{})

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		resources := Resources{}
		for k, v := range tt.Resources {
			switch k {
			case "Deployment":
				resources.IsDeployment = true
			case "Daemonset":
				resources.IsDaemonSet = true
			case "Statefulset":
				resources.IsStatefulSet = true
			}
			resources.Name = v
		}
		kubernetesOptions := &client.KubernetesOptions{
			Clientset:        clientset,
			ClusterUrl:       "",
			CurrentNamespace: tt.Namespace,
		}
		podList, err := GetResourcesPodList(kubernetesOptions, &resources, tt.Namespace)
		if err == nil && tt.Error != nil {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
		if err != nil && tt.Error == nil {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
		if err != nil && tt.Error != nil && err.Error() != tt.Error.Error() {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}

		if len(podList) != len(tt.TestPodList) {
			t.Errorf("Expected list %v found %v", tt.TestPodList, podList)
		} else {
			for i, v := range podList {
				if v != tt.TestPodList[i] {
					t.Errorf("Expected list %v found %v", tt.TestPodList, podList)
				}
			}
		}
	}
}

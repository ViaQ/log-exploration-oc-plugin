package k8sresources

import (
	"context"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetStatefulSetPodsList(t *testing.T) {
	tests := []struct {
		TestName    string
		ShouldFail  bool
		StatefulSet string
		Namespace   string
		PodList     []string
		TestPodList []string
		Error       error
	}{
		{
			"Statefulset doesn't exist",
			false,
			"dummy-statefulset",
			"openshift-logging",
			[]string{},
			[]string{},
			fmt.Errorf("stateful set \"dummy-statefulset\" not found in namespace \"openshift-logging\""),
		},
		{
			"Statefulset is present",
			false,
			"openshift-stateful",
			"openshift-logging",
			[]string{},
			[]string{"openshift-stateful"},
			nil,
		},
	}

	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "openshift-stateful",
				Namespace:   "openshift-logging",
				Annotations: map[string]string{},
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"name": "logging"},
				},
			},
		})

	clientset.CoreV1().Pods("openshift-logging").Create(context.TODO(),
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name:        "openshift-stateful",
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

		err := GetStatefulSetPodsList(clientset, &tt.PodList, tt.StatefulSet, tt.Namespace)
		if err == nil && tt.Error != nil {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
		if err != nil && tt.Error == nil {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
		if err != nil && tt.Error != nil && err.Error() != tt.Error.Error() {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}

		if len(tt.PodList) != len(tt.TestPodList) {
			t.Errorf("Expected list %v found %v", tt.TestPodList, tt.PodList)
		} else {
			for i, v := range tt.PodList {
				if v != tt.TestPodList[i] {
					t.Errorf("Expected list %v found %v", tt.TestPodList, tt.PodList)
				}
			}
		}
	}
}

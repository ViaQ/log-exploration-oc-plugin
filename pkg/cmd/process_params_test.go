package cmd

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/k8sresources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestProcessLogParameters(t *testing.T) {
	tests := []struct {
		TestName      string
		ShouldFail    bool
		TestLogParams map[string]string
		TestResources map[string]string
		Arguments     []string
		Error         error
	}{
		{
			"Logs with no parameters",
			false,
			map[string]string{},
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"deployment=openshift-deployment"},
			nil,
		},
		{
			"Logs with podname & namespace",
			false,
			map[string]string{"Podname": "openshift-logging-1234", "Namespace": "openshift-logging"},
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"deployment=openshift-deployment"},
			nil,
		},
		{
			"Logs with tail parameter",
			false,
			map[string]string{"Tail": "30m"},
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"deployment=openshift-deployment"},
			nil,
		},
		{
			"Logs with multiple parameters",
			false,
			map[string]string{
				"Podname":   "openshift-logging-1234",
				"Namespace": "openshift-logging",
				"Tail":      "30m",
				"Limit":     "5",
			},
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"deployment=openshift-deployment"},
			nil,
		},
		{
			"Logs with valid integer limit",
			false,
			map[string]string{"Limit": "5"},
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"deployment=openshift-deployment"},
			nil,
		},
		{
			"Logs with negative limit",
			false,
			map[string]string{"Limit": "-5"},
			map[string]string{"Deployment": "openshift-deployment"},
			[]string{"deployment=openshift-deployment"},
			fmt.Errorf("incorrect \"limit\" value entered, an integer value between 0 and 1000 is required"),
		},
	}

	logParameters := LogParameters{}
	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		for k, v := range tt.TestLogParams {
			switch k {
			case "Namespace":
				logParameters.Namespace = v
			case "Tail":
				logParameters.Tail = v
			case "StartTime":
				logParameters.StartTime = v
			case "EndTime":
				logParameters.EndTime = v
			case "Level":
				logParameters.Level = v
			case "Limit":
				logParameters.Limit, _ = strconv.Atoi(v)
			}
		}
		logParameters.Resources = k8sresources.Resources{}
		for k, v := range tt.TestResources {
			switch k {
			case "Deployment":
				logParameters.Resources.IsDeployment = true
			case "Daemonset":
				logParameters.Resources.IsDaemonSet = true
			case "Statefulset":
				logParameters.Resources.IsStatefulSet = true
			case "Pod":
				logParameters.Resources.IsPod = true
			}
			logParameters.Resources.Name = v
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
		kubernetesOptions := &client.KubernetesOptions{
			Clientset:        clientset,
			ClusterUrl:       "loclahost.com:8080",
			CurrentNamespace: "openshift-logging",
		}

		err := logParameters.ProcessLogParameters(kubernetesOptions, tt.Arguments)
		if err == nil && tt.Error != nil {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
		if err != nil && tt.Error == nil {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
		if err != nil && tt.Error != nil && err.Error() != tt.Error.Error() {
			t.Errorf("Expected error is %v, found %v", tt.Error, err)
		}
	}
}

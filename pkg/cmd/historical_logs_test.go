package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/k8sresources"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/logs"
	"github.com/jarcoal/httpmock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes/fake"
)

func TestExecute(t *testing.T) {
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

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		logParameters := LogParameters{}
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

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("GET", "http://log-exploration-api-route-openshift-logging.apps.com/logs/filter",
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, map[string][]string{"Logs": {
					`{"_index":"infra-000001","_type":"_doc","_id":"ODE3MjIxYjAtZDM1My00YjNmLWFiYTUtNTNjNjNkZmFjNmI2","_score":1,"_source":{"docker":{"container_id":"1128bd9f29e1846ee8351d5f397fc8c966f7b3d786f1e0596d8c918733a6082e"},"kubernetes":{"container_name":"kube-scheduler-cert-syncer","namespace_name":"openshift-kube-scheduler","pod_name":"openshift-kube-scheduler-ip-10-0-162-9.ec2.internal","container_image":"quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cf7ee380dae0dd1f3c5fb082e5b3809b0442dc9fe9e99bebb5f38b668abf54f1","container_image_id":"quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cf7ee380dae0dd1f3c5fb082e5b3809b0442dc9fe9e99bebb5f38b668abf54f1","pod_id":"c3c0585e-0b30-46b6-8897-c06eb31b520f","host":"ip-10-0-162-9.ec2.internal","master_url":"https://kubernetes.default.svc","namespace_id":"7025070c-8998-496e-a2aa-2adf38729364","namespace_labels":{"openshift_io/cluster-monitoring":"true","openshift_io/run-level":"0"},"flat_labels":["app=openshift-kube-scheduler","revision=8","scheduler=true"]},"message":"I0318 06:41:17.541040       1 certsync_controller.go:65] Syncing configmaps: []","level":"unknown","hostname":"ip-10-0-162-9.ec2.internal","pipeline_metadata":{"collector":{"ipaddr4":"10.0.162.9","inputname":"fluent-plugin-systemd","name":"fluentd","received_at":"2021-03-18T06:41:18.260559+00:00","version":"1.7.4 1.6.0"}},"@timestamp":"2021-03-18T06:41:17.541712+00:00","viaq_msg_id":"ODE3MjIxYjAtZDM1My00YjNmLWFiYTUtNTNjNjNkZmFjNmI2"}}`,
				}})
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			})

		err := logParameters.Execute(kubernetesOptions, genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, tt.Arguments)
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

func TestPrintLogs(t *testing.T) {
	tests := []struct {
		TestName   string
		ShouldFail bool
		TestLimit  int
		Error      error
	}{
		{
			"Test correct LogList",
			false,
			5,
			nil,
		},
		{
			"Empty LogList",
			false,
			5,
			fmt.Errorf("no logs present, or input parameters were invalid"),
		},
		{
			"Limit equals to 0",
			false,
			0,
			nil,
		},
		{
			"Negative limit",
			false,
			-2,
			fmt.Errorf("incorrect \"limit\" value entered, an integer value between 0 and 1000 is required"),
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		testLog := `{"_index":"infra-000001","_type":"_doc","_id":"ODE3MjIxYjAtZDM1My00YjNmLWFiYTUtNTNjNjNkZmFjNmI2","_score":1,"_source":{"docker":{"container_id":"1128bd9f29e1846ee8351d5f397fc8c966f7b3d786f1e0596d8c918733a6082e"},"kubernetes":{"container_name":"kube-scheduler-cert-syncer","namespace_name":"openshift-kube-scheduler","pod_name":"openshift-kube-scheduler-ip-10-0-162-9.ec2.internal","container_image":"quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cf7ee380dae0dd1f3c5fb082e5b3809b0442dc9fe9e99bebb5f38b668abf54f1","container_image_id":"quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:cf7ee380dae0dd1f3c5fb082e5b3809b0442dc9fe9e99bebb5f38b668abf54f1","pod_id":"c3c0585e-0b30-46b6-8897-c06eb31b520f","host":"ip-10-0-162-9.ec2.internal","master_url":"https://kubernetes.default.svc","namespace_id":"7025070c-8998-496e-a2aa-2adf38729364","namespace_labels":{"openshift_io/cluster-monitoring":"true","openshift_io/run-level":"0"},"flat_labels":["app=openshift-kube-scheduler","revision=8","scheduler=true"]},"message":"I0318 06:41:17.541040       1 certsync_controller.go:65] Syncing configmaps: []","level":"unknown","hostname":"ip-10-0-162-9.ec2.internal","pipeline_metadata":{"collector":{"ipaddr4":"10.0.162.9","inputname":"fluent-plugin-systemd","name":"fluentd","received_at":"2021-03-18T06:41:18.260559+00:00","version":"1.7.4 1.6.0"}},"@timestamp":"2021-03-18T06:41:17.541712+00:00","viaq_msg_id":"ODE3MjIxYjAtZDM1My00YjNmLWFiYTUtNTNjNjNkZmFjNmI2"}}`
		var logList []logs.LogOptions
		if strings.Compare("Empty LogList", tt.TestName) != 0 {
			logOption := logs.LogOptions{}
			err := json.Unmarshal([]byte(testLog), &logOption)
			if err != nil {
				t.Errorf("unable to fetch logs - an error occurred while unmarshalling logs")
			}
			logList = append(logList, logOption)
		}
		err := printLogs(logList, genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, tt.TestLimit, false)
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

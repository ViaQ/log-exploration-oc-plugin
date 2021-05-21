package test_e2e

import (
	"strconv"
	"testing"

	"github.com/ViaQ/log-exploration-oc-plugin/pkg/cmd"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/k8sresources"
)

func TestFetchLogs(t *testing.T) {
	tests := []struct {
		TestName      string
		ShouldFail    bool
		TestLogList   chan []string
		TestApiUrl    string
		TestLogParams map[string]string
		TestResources map[string]string
		Error         error
	}{
		{
			"Logs with no parameters",
			false,
			make(chan []string),
			"http://localhost:8080/logs/filter",
			map[string]string{},
			map[string]string{"pod": "openshift-kube-scheduler-ip-10-0-162-9.ec2.internal"},
			nil,
		},
		{
			"Logs by podname",
			false,
			make(chan []string),
			"http://localhost:8080/logs/filter",
			map[string]string{"Podname": "openshift-kube-scheduler-ip-10-0-162-9.ec2.internal"},
			map[string]string{"pod": "openshift-kube-scheduler-ip-10-0-162-9.ec2.internal"},
			nil,
		},
		{
			"Logs by given time interval",
			false,
			make(chan []string),
			"http://localhost:8080/logs/filter",
			map[string]string{"Tail": "00h30m"},
			map[string]string{"pod": "openshift-kube-scheduler-ip-10-0-162-9.ec2.internal"},
			nil,
		},
		{
			"Logs with max log limit",
			false,
			make(chan []string),
			"http://localhost:8080/logs/filter",
			map[string]string{"Limit": "5"},
			map[string]string{"pod": "openshift-kube-scheduler-ip-10-0-162-9.ec2.internal"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		logParameters := cmd.LogParameters{}
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
		go cmd.FetchLogs(tt.TestApiUrl, &logParameters, "openshift-kube-scheduler-ip-10-0-162-9.ec2.internal", tt.TestLogList)
		podLogs := <-tt.TestLogList
		if podLogs == nil {
			t.Errorf("No logs found for the pod openshift-kube-scheduler-ip-10-0-162-9.ec2.internal")
		}
	}
}

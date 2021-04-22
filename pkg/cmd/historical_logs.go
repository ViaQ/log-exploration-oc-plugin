package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	logsExample = templates.Examples(i18n.T(`
		
		# Return snapshot logs from pod openshift-apiserver-operator-849d7869ff-r94g8 with a maximum of 10 Entries
		oc historical-logs --podname=openshift-apiserver-operator-849d7869ff-r94g8 --limit=10
		
		# Return snapshot logs from namespace openshift-apiserver-operator and logging level info
		oc historical-logs --namespace=openshift-apiserver-operator --level=info
		
		# Return snapshot logs from every index without filtering
		oc historical-logs
		
		# Return snapshot of historical-logs from pods of stateful set nginx
		oc historical-logs --statefulset=nginx
		
		# Return snapshot of historical-logs from pods of deployment kibana
		oc historical-logs --deployment=kibana
		
		# Return snapshot of historical-logs from pods of daemon set fluentd
		oc historical-logs --daemonset=fluentd
		
		# Return snapshot logs in a time range between current time - 5 minutes and current time
		oc historical-logs --tail=5m
		
		# Return snapshot logs for pods in deployment log-exploration-api in the last 10 seconds
		oc historical-logs --deployment=log-exploration-api --tail=10s`))
)

type LogParameters struct {
	Namespace   string
	Podname     string
	Tail        string
	StartTime   string
	EndTime     string
	Level       string
	Limit       string
	Deployment  string
	StatefulSet string
	DaemonSet   string
}

type ResponseLogs struct {
	Logs []string
}

func NewCmdLogFilter(streams genericclioptions.IOStreams) *cobra.Command {

	o := &LogParameters{}
	cmd := &cobra.Command{
		Use:     "historical-logs [flags]",
		Short:   "View logs filtered on various parameters",
		Example: logsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Execute(streams)
			if err != nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func (o *LogParameters) AddFlags(cmd *cobra.Command) {

	cmd.Flags().StringVar(&o.Podname, "podname", "", "Filter Historical logs on a specific pod")
	cmd.Flags().StringVar(&o.Namespace, "namespace", "", "Extract Historical logs from a specific namespace")
	cmd.Flags().StringVar(&o.Tail, "tail", "", "Fetch Historical logs for the last N seconds, minutes, hours, or days")
	cmd.Flags().StringVar(&o.Level, "level", "", "Fetch Historical logs from different logging level, Example: Info,debug,Error,Unknown, etc")
	cmd.Flags().StringVar(&o.Limit, "limit", "", "Specify number of documents [logs] to be fetched. 1000 by default")
	cmd.Flags().StringVar(&o.DaemonSet, "daemonset", "", "Fetch logs from pods of a Daemon Set")
	cmd.Flags().StringVar(&o.StatefulSet, "statefulset", "", "Fetch logs from pods of a Stateful Set")
	cmd.Flags().StringVar(&o.Deployment, "deployment", "", "Fetch logs from pods of a Deployment")

}

func (o *LogParameters) Execute(streams genericclioptions.IOStreams) error {

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("kubeconfig Error: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("an error occurred while creating a kubernetes client: %v", err)
	}

	clusterUrl := config.Host
	endIndex := strings.LastIndex(clusterUrl, ":")
	startIndex := strings.Index(clusterUrl, ".") + 1
	clusterName := clusterUrl[startIndex:endIndex]

	/*Example cluster URL : http://api.sangupta-tetrh.devcluster.openshift.com:6443. The first occurrence of '.' and last occurrence of ':'
	act as start and end indices. Extract cluster name as substring using start and end Indices i.e, sangupta-tetrh.devcluster.openshift.com to build the log-exploration-api URL*/

	ApiUrl := "http://log-exploration-api-route-openshift-logging.apps." + clusterName + "/logs/filter"

	if len(o.Tail) > 0 {
		tail, err := strconv.Atoi(o.Tail[0 : len(o.Tail)-1]) //extract numeric value. For example, extract 50 from 50s or 10 from 10m
		if err != nil {
			return fmt.Errorf("an invalid \"tail\" value was entered: %v", err)
		}

		timeUnit := o.Tail[len(o.Tail)-1] //Last character (time unit) is 's'(seconds),'m'(minutes),'h'(hours),'d'(days)
		endTime := time.Now().UTC()
		var startTime time.Time

		switch timeUnit {
		case 's':
			startTime = endTime.Add(time.Duration(-tail) * time.Second).UTC()
		case 'm':
			startTime = endTime.Add(time.Duration(-tail) * time.Minute).UTC()
		case 'h':
			startTime = endTime.Add(time.Duration(-tail) * time.Hour).UTC()
		case 'd':
			startTime = endTime.Add(time.Duration(-tail) * time.Hour * 24)
		default:
			return fmt.Errorf("invalid time unit entered in \"tail\". please enter s, m, h, or d as time unit")
		}

		o.StartTime = startTime.UTC().Format(time.RFC3339Nano)
		o.EndTime = endTime.UTC().Format(time.RFC3339Nano)
	}

	var podList []string

	if len(o.Deployment) > 0 {
		err = GetDeploymentPodsList(clientset, o.Deployment, o.Namespace, &podList)
		if err != nil {
			return err
		}
	}

	if len(o.DaemonSet) > 0 {
		err = GetDaemonSetPodsList(clientset, o.DaemonSet, o.Namespace, &podList)
		if err != nil {
			return err
		}
	}

	if len(o.StatefulSet) > 0 {
		err = GetStatefulSetPodsList(clientset, o.StatefulSet, o.Namespace, &podList)
		if err != nil {
			return err
		}
	}

	var logList []string
	if len(podList) > 0 {
		for _, pod := range podList {

			o.Podname = pod
			response, err := o.makeHttpRequest(ApiUrl)

			if err != nil {
				return err
			}

			responseBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %v", err)
			}

			err = getLogs(responseBody, &logList)
			if err != nil {
				return err
			}
		}

		err = printLogs(logList, streams, o.Limit)
		if err != nil {
			return err
		}

	} else {

		response, err := o.makeHttpRequest(ApiUrl)
		if err != nil {
			return err
		}
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}

		err = getLogs(responseBody, &logList)
		if err != nil {
			return err
		}

		err = printLogs(logList, streams, o.Limit)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *LogParameters) makeHttpRequest(baseUrl string) (*http.Response, error) {

	var urlBuilder strings.Builder
	urlBuilder.WriteString(baseUrl)
	numParameters := 0
	if len(o.Podname) > 0 {
		urlBuilder.WriteString("?")
		urlBuilder.WriteString("podname=" + o.Podname)
		numParameters = numParameters + 1
	}
	if len(o.Namespace) > 0 {
		if numParameters == 0 {
			urlBuilder.WriteString("?")
		} else {
			urlBuilder.WriteString("&")
		}
		urlBuilder.WriteString("namespace=" + o.Namespace)
		numParameters = numParameters + 1
	}
	if len(o.Limit) > 0 {
		if numParameters == 0 {
			urlBuilder.WriteString("?")
		} else {
			urlBuilder.WriteString("&")
		}
		urlBuilder.WriteString("maxlogs=" + o.Limit)
		numParameters = numParameters + 1
	}
	if len(o.Level) > 0 {
		if numParameters == 0 {
			urlBuilder.WriteString("?")
		} else {
			urlBuilder.WriteString("&")
		}
		urlBuilder.WriteString("level=" + o.Level)
		numParameters = numParameters + 1
	}
	if len(o.StartTime) > 0 && len(o.EndTime) > 0 {
		if numParameters == 0 {
			urlBuilder.WriteString("?")
		} else {
			urlBuilder.WriteString("&")
		}
		urlBuilder.WriteString("starttime=" + o.StartTime)
		urlBuilder.WriteString("&finishtime=" + o.EndTime)
		numParameters = numParameters + 1
	}
	response, err := http.Get(urlBuilder.String())

	if err != nil {
		return nil, fmt.Errorf("http request failed: %v", err)
	}
	return response, nil
}

func printLogs(logList []string, streams genericclioptions.IOStreams, limit string) error {

	if len(logList) == 0 {
		_, err := fmt.Fprintf(streams.Out, "No Logs Present, or logs from invalid resources were requested")
		if err != nil {
			return fmt.Errorf("an error occurred while printing logs: %v", err)
		}
		return nil
	}

	var maxLogs int
	var err error

	if len(limit) > 0 {
		maxLogs, err = strconv.Atoi(limit)
		if err != nil {
			return fmt.Errorf("incorrect \"limit\" value entered")
		}
	}

	for logCount, log := range logList {

		if len(limit) > 0 && logCount > maxLogs {
			return nil
		}
		_, err := fmt.Fprintf(streams.Out, log+"\n")
		if err != nil {
			return fmt.Errorf("an error occurred while printing logs: %v", err)
		}
	}
	return nil
}

func getLogs(responseBody []byte, logList *[]string) error {

	jsonResponse := &ResponseLogs{}
	err := json.Unmarshal(responseBody, &jsonResponse)
	if err != nil {
		return fmt.Errorf("an error occurred while unmarshalling JSON response: %v", err)
	}

	for _, log := range jsonResponse.Logs {

		logOption := LogOptions{}
		err := json.Unmarshal([]byte(log), &logOption)

		if err != nil {
			return nil //no logs present
		}

		if len(logOption.Source.Message) > 0 {
			*logList = append(*logList, logOption.Source.Message)
		}
	}

	return nil
}

func GetDaemonSetPodsList(clientset *kubernetes.Clientset, targetDaemonSet string, namespace string, podList *[]string) error {

	var requiredDaemonSet appsv1.DaemonSet
	daemonsets, _ := clientset.AppsV1().DaemonSets(namespace).List(context.Background(), metav1.ListOptions{})

	requiredDaemonSetFound := false
	for _, daemonset := range daemonsets.Items {
		if daemonset.ObjectMeta.Name == targetDaemonSet {
			requiredDaemonSetFound = true
			requiredDaemonSet = daemonset
			break
		}
	}

	if !requiredDaemonSetFound {
		return fmt.Errorf("daemon set \"%v\" not found", targetDaemonSet)
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

func GetStatefulSetPodsList(clientset *kubernetes.Clientset, targetStatefulSet string, namespace string, podList *[]string) error {

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
		return fmt.Errorf("stateful set \"%v\" not found", targetStatefulSet)

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

func GetDeploymentPodsList(clientset *kubernetes.Clientset, targetDeployment string, namespace string, podList *[]string) error {

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
		return fmt.Errorf("deployment \"%v\" not found", targetDeployment)
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

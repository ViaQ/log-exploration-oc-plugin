package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/k8sresources"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/logs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"net/http"
	"strconv"
	"strings"
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
		oc historical-logs --deployment=log-exploration-api --tail=10s

		# Return snapshot logs for pods in deployment log-exploration-api in namespace openshift-logging
		oc historical-logs --deployment=log-exploration-api --namespace=openshift-logging`))
)

type ResponseLogs struct {
	Logs []string
}

type LogParameters struct {
	Namespace string
	Podname   string
	Tail      string
	StartTime string
	EndTime   string
	Level     string
	Limit     string
	Prefix    bool
	k8sresources.Resources
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
	cmd.Flags().StringVar(&o.Resources.DaemonSet, "daemonset", "", "Fetch logs from pods of a Daemon Set")
	cmd.Flags().StringVar(&o.Resources.StatefulSet, "statefulset", "", "Fetch logs from pods of a Stateful Set")
	cmd.Flags().StringVar(&o.Resources.Deployment, "deployment", "", "Fetch logs from pods of a Deployment")
	cmd.Flags().BoolVar(&o.Prefix, "prefix", false, "Prefix each log with the log source (pod name and container name)")
}

func (o *LogParameters) Execute(streams genericclioptions.IOStreams) error {

	kubernetesOptions, err := client.KubernetesClient()
	if err != nil {
		return err
	}

	err = o.ProcessLogParameters()

	if err != nil {
		return err
	}

	var podList []string

	podList, err = k8sresources.GetResourcesPodList(kubernetesOptions, &o.Resources, o.Namespace)

	if err != nil {
		return err
	}

	endIndex := strings.LastIndex(kubernetesOptions.ClusterUrl, ":")
	startIndex := strings.Index(kubernetesOptions.ClusterUrl, ".") + 1
	clusterName := kubernetesOptions.ClusterUrl[startIndex:endIndex]

	/*Example cluster URL : http://api.sangupta-tetrh.devcluster.openshift.com:6443. The first occurrence of '.' and last occurrence of ':'
	act as start and end indices. Extract cluster name as substring using start and end Indices i.e, sangupta-tetrh.devcluster.openshift.com to build the log-exploration-api URL*/

	baseUrl := "http://log-exploration-api-route-openshift-logging.apps." + clusterName + "/logs/filter"

	var logList []string

	if len(podList) > 0 {
		for _, pod := range podList {

			o.Podname = pod

			err := fetchLogs(&logList, baseUrl, o)
			if err != nil {
				return err
			}
		}

		err = printLogs(logList, streams, o.Limit)
		if err != nil {
			return err
		}

	} else {

		err = fetchLogs(&logList, baseUrl, o)
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

func fetchLogs(logList *[]string, baseUrl string, logParameters *LogParameters) error {

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return fmt.Errorf("http request failed: %v", err)
	}

	query := req.URL.Query()
	query.Add("podname", logParameters.Podname)
	query.Add("namespace", logParameters.Namespace)
	query.Add("starttime", logParameters.StartTime)
	query.Add("finishtime", logParameters.EndTime)
	query.Add("maxlogs", logParameters.Limit)
	query.Add("level", logParameters.Level)
	req.URL.RawQuery = query.Encode()

	response, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to get http response %v", err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	err = response.Body.Close()

	if err != nil {
		return fmt.Errorf("an error occurred while attempting to close response body %v", err)
	}

	jsonResponse := &ResponseLogs{}
	err = json.Unmarshal(responseBody, &jsonResponse)
	if err != nil {
		return fmt.Errorf("an error occurred while unmarshalling JSON response: %v", err)
	}

	for _, log := range jsonResponse.Logs {

		logOption := logs.LogOptions{}
		err := json.Unmarshal([]byte(log), &logOption)

		if err != nil {
			return fmt.Errorf("no logs present, or input parameters were invalid")
		}

		if len(logOption.Source.Message) > 0 {
			if logParameters.Prefix && len(logOption.Source.Kubernetes.PodName) > 0 && len(logOption.Source.Kubernetes.ContainerName) > 0 {
				*logList = append(*logList, "pod/"+logOption.Source.Kubernetes.PodName+"/"+logOption.Source.Kubernetes.ContainerName+"   "+logOption.Source.Message)
			} else {
				*logList = append(*logList, logOption.Source.Message)
			}
		}
	}

	return nil
}

func printLogs(logList []string, streams genericclioptions.IOStreams, limit string) error {

	if len(logList) == 0 {
		return fmt.Errorf("no logs present, or input parameters were invalid")
	}

	var maxLogs int
	var err error

	if len(limit) > 0 {
		maxLogs, err = strconv.Atoi(limit)
		if err != nil || maxLogs < 0 || maxLogs > 1000 {
			return fmt.Errorf("incorrect \"limit\" value entered, an integer value between 0 and 1000 is required")
		}
	}

	for logCount, log := range logList {

		if len(limit) > 0 && logCount >= maxLogs {
			return nil
		}
		_, err := fmt.Fprintf(streams.Out, log+"\n")
		if err != nil {
			return fmt.Errorf("an error occurred while printing logs: %v", err)
		}
	}
	return nil
}

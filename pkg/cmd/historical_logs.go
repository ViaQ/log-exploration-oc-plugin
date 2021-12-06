package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/constants"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/k8sresources"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/logs"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	logsExample = templates.Examples(i18n.T(`
		# Return snapshot historical-logs from pod openshift-apiserver-operator-849d7869ff-r94g8 with a maximum of 10 log extries
		oc historical-logs podname=openshift-apiserver-operator-849d7869ff-r94g8 --limit=10
		
		# Return snapshot of historical-logs from pods of stateful set prometheus from namespace openshift-apiserver-operator and logging level info
		oc historical-logs statefulset=prometheus --namespace=openshift-apiserver-operator --level=info
		
		# Return snapshot of historical-logs from pods of stateful set nginx in the current namespace with pod name and container name as log prefix
		oc historical-logs statefulset=nginx --prefix=true
		
		# Return snapshot of historical-logs from pods of deployment kibana in the namespace openshift-logging with a maximum of 100 log entries
		oc historical-logs deployment=kibana --namespace=openshift-logging --limit=100
		
		# Return snapshot of historical-logs from pods of daemon set fluentd in the current namespace
		oc historical-logs daemonset=fluentd
		
		# Return snapshot logs of pods in deployment cluster-logging-operator in a time range between current time - 5 minutes and current time
		oc historical-logs deployment=cluster-logging-operator --tail=5m
		
		# Return snapshot logs for pods in deployment log-exploration-api in the last 10 seconds
		oc historical-logs deployment=log-exploration-api --tail=10s`))
)

type ResponseLogs struct {
	Logs []string
}

type LogParameters struct {
	Namespace string
	Tail      string
	StartTime string
	EndTime   string
	Level     string
	Limit     int
	Prefix    bool
	k8sresources.Resources
}

func NewCmdLogFilter(streams genericclioptions.IOStreams) *cobra.Command {

	o := &LogParameters{}

	cmd := &cobra.Command{
		Use:     "historical-logs [resource-type]=[resource-name] [flags]",
		Short:   "View logs filtered on various parameters",
		Example: logsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			kubernetesOptions, err := client.KubernetesClient()
			if err != nil {
				return err
			}
			err = o.Execute(kubernetesOptions, streams, args)
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

	cmd.Flags().StringVar(&o.Namespace, "namespace", "", "Extract Historical logs from a specific namespace")
	cmd.Flags().StringVar(&o.Tail, "tail", "", "Fetch Historical logs for the last N seconds, minutes, hours, or days")
	cmd.Flags().StringVar(&o.Level, "level", "", "Fetch Historical logs from different logging level, Example: Info,debug,Error,Unknown, etc")
	cmd.Flags().IntVar(&o.Limit, "limit", constants.LimitUpperBound, "Specify number of documents [logs] to be fetched")
	cmd.Flags().BoolVar(&o.Prefix, "prefix", false, "Prefix each log with the log source (pod name and container name)")
}

func (o *LogParameters) Execute(kubernetesOptions *client.KubernetesOptions, streams genericclioptions.IOStreams, args []string) error {
	err := o.ProcessLogParameters(kubernetesOptions, args)

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
	baseUrl := "http://log-exploration-api-route-openshift-logging.apps." + clusterName + "/logs"

	podLogsCh := make(chan []logs.LogOptions)
	var logList []logs.LogOptions
	for _, pod := range podList {
		go FetchLogs(baseUrl, o, pod, podLogsCh)
	}

	for index := 0; index < len(podList); index++ {
		podLogs := <-podLogsCh
		logList = append(logList, podLogs...)
	}

	sort.Slice(logList, func(index1, index2 int) bool {
		return logList[index1].Source.Timestamp.String() > logList[index2].Source.Timestamp.String()
	})

	err = printLogs(logList, streams, o.Limit, o.Prefix)
	if err != nil {
		return err
	}

	return nil

}

func FetchLogs(baseUrl string, logParameters *LogParameters, podname string, podLogsCh chan<- []logs.LogOptions) {

	req, err := http.NewRequest("GET", baseUrl, nil)

	if err != nil {
		fmt.Printf("unable to fetch logs of pod %s - http request failed: %v\n", podname, err)
		podLogsCh <- nil
		return
	}

	out, err := exec.Command("bash", "-c", "oc whoami --show-token").Output()
	if err != nil {
		fmt.Println("Error", err)
	}
	out = out[:len(out)-1]
	var bearer = "`Bearer " + string(out) + "`"
	req.Header.Set("Authorization", bearer)

	query := req.URL.Query()
	query.Add("/pod/", podname)
	query.Add("/namespace/", logParameters.Namespace)
	query.Add("/starttime/", logParameters.StartTime)
	query.Add("/finishtime/", logParameters.EndTime)
	query.Add("/maxlogs/", strconv.Itoa(logParameters.Limit))
	query.Add("/level/", logParameters.Level)
	req.URL.RawQuery = query.Encode()
	fmt.Println(req.URL)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("unable to fetch logs of pod %s - failed to get http response %v\n", podname, err)
		podLogsCh <- nil
		return
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("unable to fetch logs of pod %s - failed to read response: %v\n", podname, err)
		podLogsCh <- nil
		return
	}

	err = response.Body.Close()

	if err != nil {
		fmt.Printf("unable to fetch logs of pod %s - an error occurred while attempting to close response body %v\n", podname, err)
		podLogsCh <- nil
		return
	}

	jsonResponse := &ResponseLogs{}
	err = json.Unmarshal(responseBody, &jsonResponse)
	if err != nil {
		fmt.Printf("unable to fetch logs of pod %s - an error occurred while unmarshalling JSON response: %v\n", podname, err)
		podLogsCh <- nil
		return
	}

	var logList []logs.LogOptions
	for _, log := range jsonResponse.Logs {
		logOption := logs.LogOptions{}
		err := json.Unmarshal([]byte(log), &logOption)

		if err != nil {
			fmt.Printf("unable to fetch logs of pod %s - no logs present, or input parameters were invalid %v\n", podname, err)
			podLogsCh <- nil
			return
		}
		logList = append(logList, logOption)
	}

	podLogsCh <- logList
}

func printLogs(logList []logs.LogOptions, streams genericclioptions.IOStreams, limit int, prefix bool) error {

	if len(logList) == 0 {
		return fmt.Errorf("no logs present, or input parameters were invalid")
	}

	for logCount, log := range logList {
		if limit < 0 {
			return fmt.Errorf("incorrect \"limit\" value entered, an integer value between 0 and 1000 is required")
		}
		if logCount >= limit {
			return nil
		}

		if len(log.Source.Message) > 0 {
			if prefix && len(log.Source.Kubernetes.PodName) > 0 && len(log.Source.Kubernetes.ContainerName) > 0 {
				_, err := fmt.Fprintf(streams.Out, "pod/"+log.Source.Kubernetes.PodName+"/"+log.Source.Kubernetes.ContainerName+"   "+log.Source.Message+"\n")
				if err != nil {
					return fmt.Errorf("an error occurred while printing logs: %v", err)
				}
			} else {
				_, err := fmt.Fprintf(streams.Out, log.Source.Message+"\n")
				if err != nil {
					return fmt.Errorf("an error occurred while printing logs: %v", err)
				}
			}

		}

	}

	return nil
}

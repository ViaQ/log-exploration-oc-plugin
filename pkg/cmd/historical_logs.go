package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var
(
	logsExample = templates.Examples(i18n.T(`
		# Return snapshot logs from pod openshift-apiserver-operator-849d7869ff-r94g8 with a maximum of 10 Entries
		kubectl historical-logs --podname=openshift-apiserver-operator-849d7869ff-r94g8 --maxlogs=10
		# Return snapshot logs from namespace openshift-apiserver-operator and logging level info
		kubectl historical-logs --namespace=openshift-apiserver-operator --level=info
		# Return snapshot logs from every index without filtering
		kubectl historical-logs
		# Return snapshot of historical-logs from Infrastructure index
		kubectl historical-logs --index=infra-000001
		# Return snapshot logs in a time range between start and end times
		kubectl historical-logs --starttime=2021-03-09T03:40:00.163677339Z --finishtime=2021-03-09T03:50:00.163677339Z
		# Return snapshot logs in a time range between start and end times for the infrastructure index
		kubectl historical-logs --starttime=2021-03-09T03:40:00.163677339Z --finishtime=2021-03-09T03:50:00.163677339Z --index=infra-000001
		# Return snapshot a maximum of 20 historical-logs`))
)

type LogParameters struct {
	Namespace string `json:"namespace"`
	Index string `json:"index"`
	Podname string `json:"podname"`
	StartTime string `json:"starttime"`
	FinishTime string `json:"finishtime"`
	Level string `json:"level"`
	MaxLogs string `json:"maxlogs"`
}

type ResponseLogs struct{
	Logs []string `json:"Logs"`
}


func NewCmdLogFilter(streams genericclioptions.IOStreams) *cobra.Command {

	o := &LogParameters{}
	cmd := &cobra.Command{
		Use:     "historical-logs [flags]",
		Short:   "View logs filtered on various parameters",
		Example: logsExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Execute(streams)
			if err!=nil {
				return err
			}
			return nil
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func (o *LogParameters)AddFlags(cmd *cobra.Command) {

	cmd.Flags().StringVar(&o.Podname, "podname", "", "Filter Historical logs on a specific pod name.")
	cmd.Flags().StringVar(&o.Namespace, "namespace", "", "Extract Historical logs from a specific namespace")
	cmd.Flags().StringVar(&o.StartTime, "starttime", "", "Fetch Historical logs from a particular timestamp")
	cmd.Flags().StringVar(&o.FinishTime, "finishtime", "", "Fetch Historical logs till a timestamp")
	cmd.Flags().StringVar(&o.Level, "level", "", "Fetch Historical logs from different logging level, Example: Info,debug,Error,Unknown, etc")
	cmd.Flags().StringVar(&o.MaxLogs, "maxlogs", "", "Specify number of documents [logs] to be fetched. 1000 by default")
	cmd.Flags().StringVar(&o.Index, "index", "", "Specify Index to filter by [infra-000001/app-000001/audit-000001")

}


func (o *LogParameters) Execute(streams genericclioptions.IOStreams) error{

	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("Kubeconfig Error: ",err)
	}

	payload := new(bytes.Buffer)
	err=json.NewEncoder(payload).Encode(o)

	if err!=nil{
		return fmt.Errorf("An Error Occurred while encoding JSON: ",err)
	}

	clusterUrl := config.Host
	endIndex := strings.LastIndex(clusterUrl,":")
	startIndex := strings.Index(clusterUrl,".")+1
	clusterName := clusterUrl[startIndex:endIndex]

	/*Example cluster URL : http://api.sangupta-tetrh.devcluster.openshift.com:6443. The first occurrence of '.' and last occurrence of ':'
	 act as start and end indices. Extract cluster name as substring using start and end Indices i.e, sangupta-tetrh.devcluster.openshift.com to build the log-exploration-api URL*/

	ApiUrl := "http://log-exploration-api-route-openshift-logging.apps."+clusterName+"/logs/filter"

	req, err := http.NewRequest("GET", ApiUrl, payload)

	if err!=nil {
		return fmt.Errorf("Request Failed: ",err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Response Error: ",err)
	}

	responseBody,err := ioutil.ReadAll(res.Body)
	if err!=nil {
		return fmt.Errorf("Failed to read Response: ",err)
	}

	jsonResponse := &ResponseLogs{}
	json.Unmarshal(responseBody,&jsonResponse)

	for i:=0;i<len(jsonResponse.Logs);i++{

		log := LogOptions{}
		err = json.Unmarshal([]byte(jsonResponse.Logs[i]),&log)

		if err!=nil {
			return fmt.Errorf("An Error Occurred while decoding response: ",err)
		}

		if len(log.Source.Message)>0 {
			fmt.Fprintf(streams.Out, log.Source.Message+"\n")
		}

	}

	return nil
}

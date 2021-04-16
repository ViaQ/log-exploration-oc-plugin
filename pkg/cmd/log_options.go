package cmd

import "time"

type LogOptions struct {
	ID     string `json:"_id"`
	Index  string `json:"_index"`
	Score  int    `json:"_score"`
	Source struct {
		Timestamp time.Time `json:"@timestamp"`
		Docker    struct {
			ContainerID string `json:"container_id"`
		} `json:"docker"`
		Hostname   string `json:"hostname"`
		Kubernetes struct {
			ContainerImage   string   `json:"container_image"`
			ContainerImageID string   `json:"container_image_id"`
			ContainerName    string   `json:"container_name"`
			FlatLabels       []string `json:"flat_labels"`
			Host             string   `json:"host"`
			MasterURL        string   `json:"master_url"`
			NamespaceID      string   `json:"namespace_id"`
			NamespaceLabels  struct {
				OpenshiftIoClusterMonitoring string `json:"openshift_io/cluster-monitoring"`
				OpenshiftIoRunLevel          string `json:"openshift_io/run-level"`
			} `json:"namespace_labels"`
			NamespaceName string `json:"namespace_name"`
			PodID         string `json:"pod_id"`
			PodName       string `json:"pod_name"`
		} `json:"kubernetes"`
		Level            string `json:"level"`
		Message          string `json:"message"`
		PipelineMetadata struct {
			Collector struct {
				Inputname  string    `json:"inputname"`
				Ipaddr4    string    `json:"ipaddr4"`
				Name       string    `json:"name"`
				ReceivedAt time.Time `json:"received_at"`
				Version    string    `json:"version"`
			} `json:"collector"`
		} `json:"pipeline_metadata"`
		ViaqMsgID string `json:"viaq_msg_id"`
	} `json:"_source"`
	Type string `json:"_type"`
}

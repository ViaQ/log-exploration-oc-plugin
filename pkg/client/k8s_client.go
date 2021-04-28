package client

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type KubernetesOptions struct {
	Clientset  kubernetes.Interface
	ClusterUrl string
}

func KubernetesClient() (*KubernetesOptions, error) {

	kubernetesOptions := &KubernetesOptions{}
	kubeconfig := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig Error: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("an error occurred while creating kubernetes client: %v", err)
	}
	kubernetesOptions.Clientset = clientset
	kubernetesOptions.ClusterUrl = config.Host
	return kubernetesOptions, nil
}

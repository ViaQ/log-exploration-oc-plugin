package k8sresources

type Resources struct {
	IsDeployment  bool
	IsDaemonSet   bool
	IsStatefulSet bool
	IsPod         bool
	Name          string
}

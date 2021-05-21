package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/constants"
)

func (o *LogParameters) ProcessLogParameters(kubernetesOptions *client.KubernetesOptions, args []string) error {

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

	if o.Limit < constants.LimitLowerBound || o.Limit > constants.LimitUpperBound {
		return fmt.Errorf("incorrect \"limit\" value entered, an integer value between %d and %d is required", constants.LimitLowerBound, constants.LimitUpperBound)
	}

	if len(o.Namespace) == 0 {
		o.Namespace = kubernetesOptions.CurrentNamespace
	}

	if len(args) != 1 {
		return fmt.Errorf("one of deployment/daemonset/statefulset/podname required as argument in the format - [resource-type]=[resource-name]")
	}

	resourceTypeNameSplit := strings.Split(args[0], "=") //example command- oc historical-logs deployment=deployment1 hence, splitting on "=" to extract resource and name

	if len(resourceTypeNameSplit) != 2 {
		return fmt.Errorf("invalid format. [resource-type]=[resource-name] required as argument")
	}

	resourceType := resourceTypeNameSplit[0]
	resourceName := resourceTypeNameSplit[1]

	switch resourceType {
	case constants.Deployment:
		o.Resources.IsDeployment = true
		o.Resources.Name = resourceName
	case constants.DaemonSet:
		o.Resources.IsDaemonSet = true
		o.Resources.Name = resourceName
	case constants.StatefulSet:
		o.Resources.IsStatefulSet = true
		o.Resources.Name = resourceName
	case constants.Podname:
		o.Resources.IsPod = true
		o.Resources.Name = resourceName
	default:
		return fmt.Errorf("logs for invalid resource type \"%s\" requested", resourceType)
	}
	return nil
}

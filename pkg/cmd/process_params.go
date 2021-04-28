package cmd

import (
	"fmt"
	"github.com/ViaQ/log-exploration-oc-plugin/pkg/client"
	"strconv"
	"time"
)

func (o *LogParameters) ProcessLogParameters() error {

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
	return nil
}

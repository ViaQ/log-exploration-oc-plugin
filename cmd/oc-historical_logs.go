package main

import (
	cmd "github.com/ViaQ/log-exploration-oc-plugin/pkg/cmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
)

func main() {

	root := cmd.NewCmdLogFilter(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

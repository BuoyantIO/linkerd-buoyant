package cmd

import (
	"fmt"
	"io"
)

type config struct {
	kubecontext  string
	kubeconfig   string
	verbose      bool
	bcloudServer string
	stdout       io.Writer
	stderr       io.Writer
}

func (c *config) printVerbosef(format string, a ...interface{}) {
	if c.verbose {
		fmt.Fprintf(c.stderr, format+"\n", a...)
	}
}

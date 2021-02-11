package main

import (
	"os"

	"github.com/buoyantio/linkerd-buoyant/cmd"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		os.Exit(1)
	}
}

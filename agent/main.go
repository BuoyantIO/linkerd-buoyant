package main

import (
	"fmt"
	"os"

	"github.com/buoyantio/linkerd-buoyant/agent/cmd/agent"
	"github.com/buoyantio/linkerd-buoyant/agent/cmd/registrator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected a subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "registrator":
		registrator.Main(os.Args[2:])
	case "agent":
		agent.Main(os.Args[2:])
	default:
		fmt.Printf("unknown subcommand: %s", os.Args[1])
		os.Exit(1)
	}
}

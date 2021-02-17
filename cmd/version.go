package cmd

import (
	"context"
	"fmt"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/buoyantio/linkerd-buoyant/pkg/version"
	"github.com/spf13/cobra"
)

type versionConfig struct {
	*config
	short bool
	cli   bool
}

func newCmdVersion(cfg *config) *cobra.Command {
	versionCfg := &versionConfig{config: cfg}

	cmd := &cobra.Command{
		Use:   "version [flags]",
		Args:  cobra.NoArgs,
		Short: "Print the CLI and Agent version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.New(cfg.kubeconfig, cfg.kubecontext, cfg.bcloudServer)
			if err != nil {
				return err
			}

			return versionCmd(cmd.Context(), versionCfg, client)
		},
	}

	cmd.Flags().BoolVar(&versionCfg.short, "short", false, "Print the version number(s) only, with no additional output")
	cmd.Flags().BoolVar(&versionCfg.cli, "cli", false, "Print the CLI version only")

	return cmd
}

func versionCmd(
	ctx context.Context, cfg *versionConfig, client k8s.Client,
) error {
	if cfg.short {
		fmt.Fprintf(cfg.stdout, "%s\n", version.Version)
	} else {
		fmt.Fprintf(cfg.stdout, "CLI version:   %s\n", version.Version)
	}

	if cfg.cli {
		return nil
	}

	agent, err := client.Agent(ctx)
	if err != nil {
		fmt.Fprintf(cfg.stderr, "Failed to get Agent version: %s\n", err)
		return err
	}

	agentVersion := "not found"
	if agent != nil {
		agentVersion = agent.Version
	}

	if cfg.short {
		fmt.Fprintf(cfg.stdout, "%s\n", agentVersion)
	} else {
		fmt.Fprintf(cfg.stdout, "Agent version: %s\n", agentVersion)
	}

	return nil
}

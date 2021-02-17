package cmd

import (
	"context"
	"fmt"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/buoyantio/linkerd-buoyant/pkg/version"
	"github.com/spf13/cobra"
)

type versionConfig struct {
	config
	short   bool
	cliOnly bool
}

func newCmdVersion(cfg config) *cobra.Command {
	versionCfg := versionConfig{config: cfg}

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
	cmd.Flags().BoolVar(&versionCfg.cliOnly, "client", false, "Print the client version only")

	return cmd
}

func versionCmd(
	ctx context.Context, cfg versionConfig, client k8s.Client,
) error {
	if cfg.short {
		fmt.Fprintf(cfg.stdout, "%s\n", version.Version)
	} else {
		fmt.Fprintf(cfg.stdout, "CLI version:   %s\n", version.Version)
	}

	if cfg.cliOnly {
		return nil
	}

	agent, err := client.Agent(ctx)
	if err != nil {
		fmt.Fprintf(cfg.stderr, "Failed to get server version: %s\n", err)
	}
	if cfg.short {
		fmt.Fprintf(cfg.stdout, "%s\n", agent.Version)
	} else {
		fmt.Fprintf(cfg.stdout, "Agent version: %s\n", agent.Version)
	}

	return nil
}

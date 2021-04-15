package cmd

import (
	"context"
	"fmt"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func newCmdDashboard(cfg *config) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "dashboard [flags]",
		Args:  cobra.NoArgs,
		Short: "Opens the Bcloud Dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.New(cfg.kubeconfig, cfg.kubecontext, cfg.bcloudServer)
			if err != nil {
				return err
			}

			return dashboardCmd(cmd.Context(), cfg, client, browser.OpenURL)
		},
	}

	return cmd
}

func dashboardCmd(
	ctx context.Context, cfg *config, client k8s.Client, openURL openURL,
) error {

	agent, err := client.Agent(ctx)
	if err != nil {
		cfg.printVerbosef("Failed to get Agent: %s\n", err)
		return nil
	}

	if agent == nil {
		fmt.Fprint(cfg.stdout, "Agent not installed, run linkerd-buoyant install\n")
		return nil
	}

	return openURL(cfg.bcloudServer)
}

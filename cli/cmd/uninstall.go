package cmd

import (
	"context"
	"fmt"

	"github.com/buoyantio/linkerd-buoyant/cli/pkg/k8s"
	"github.com/spf13/cobra"
)

func newCmdUninstall(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Kubernetes manifest to uninstall the Buoyant Cloud agent",
		Long: `Output Kubernetes manifest to uninstall the Buoyant Cloud agent.

This command provides all Kubernetes namespace-scoped and cluster-scoped
resources (e.g Namespace, RBACs, etc.) necessary to uninstall the Buoyant Cloud
Agent.`,
		Example: `  # Default uninstall.
  linkerd buoyant uninstall | kubectl delete -f -

  # Unnstall from a specific cluster
  linkerd buoyant --context test-cluster uninstall | kubectl delete --context test-cluster -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.New(cfg.kubeconfig, cfg.kubecontext)
			if err != nil {
				return err
			}

			return uninstall(cmd.Context(), cfg, client)
		},
	}

	return cmd
}

func uninstall(ctx context.Context, cfg *config, client k8s.Client) error {
	resources, err := client.Resources(ctx)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		fmt.Fprintf(cfg.stderr, "No linkerd-buoyant resources found on cluster.\n")
		return nil
	}

	cfg.printVerbosef("Found %d resources for deletion", len(resources))

	// retrieve agent prior to deletion
	agent, _ := client.Agent(ctx)

	// print out all linkerd-buoyant resources to be deleted
	if len(resources) > 0 {
		for _, r := range resources {
			fmt.Fprintf(cfg.stdout, "%s---\n", r)
		}
		fmt.Fprintf(cfg.stderr, "\n")
	}

	// if agent is present, output its name and command for posterity
	if agent != nil {
		fmt.Fprintf(cfg.stderr, "Agent manifest will remain available via:\nlinkerd buoyant install --cluster-name=%s\n\n", agent.Name)
	}

	return nil
}

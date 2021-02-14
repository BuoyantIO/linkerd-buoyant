package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/spf13/cobra"
)

func newCmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Kubernetes resources to uninstall the linkerd-buoyant extension",
		Long: `Output Kubernetes resources to uninstall the linkerd-buoyant extension.

This command provides all Kubernetes namespace-scoped and cluster-scoped
resources (e.g services, deployments, RBACs, etc.) necessary to uninstall the
linkerd-buoyant extension.`,
		Example: `  # Default uninstall.
  linkerd buoyant uninstall | kubectl delete -f -

  # Unnstall from a specific cluster
  linkerd buoyant --context test-cluster uninstall | kubectl delete --context test-cluster -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstall(cmd.Context())
		},
	}

	return cmd
}

func uninstall(ctx context.Context) error {
	client, err := k8s.New(kubeconfig, kubecontext)
	if err != nil {
		return err
	}

	resources, err := k8s.Resources(ctx, client)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		fmt.Fprintf(os.Stderr, "No linkerd-buoyant resources found on cluster.\n")
		return nil
	}

	printVerbosef("Found %d resources for deletion", len(resources))

	// retrieve agent prior to deletion
	agent, _ := k8s.GetAgent(ctx, client, bcloudServer)

	// print out all linkerd-buoyant resources to be deleted
	for _, r := range resources {
		fmt.Fprintf(os.Stdout, "%s---\n", r)
	}

	// if agent is present, output its name and URL for posterity
	if agent != nil {
		fmt.Fprintf(os.Stderr, "Agent manifest will remain available at:\n%s\n", agent.URL)
	}

	return nil
}

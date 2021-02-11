package cmd

import (
	"github.com/spf13/cobra"
)

func install() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Buoyant Cloud Agent manifest for installation",
		Long: `Output Buoyant Cloud Agent manifest for installation.

This command provides the Kubernetes configs necessary to install the Buoyant
Cloud Agent.

If an agent is not already present on the current cluster, this command
redirects the user to Buoyant Cloud to set up a new agent. Once the new agent is
set up, this command will output the agent manifest.

If an agent is already present, this command retrieves an updated manifest from
Buoyant Cloud and outputs it.`,
		Example: `  # Default install.
  linkerd buoyant install | kubectl apply -f -

  # Install onto a specific cluster
  linkerd buoyant --context test-cluster install | kubectl apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}

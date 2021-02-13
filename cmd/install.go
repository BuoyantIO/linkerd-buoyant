package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/spf13/cobra"
)

func newCmdInstall() *cobra.Command {
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
  linkerd buoyant --context test-cluster install | kubectl --context test-cluster apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(cmd.Context())
		},
	}

	return cmd
}

func install(ctx context.Context) error {
	client, currentContext, err := k8s.New(kubeconfig, kubecontext)
	if err != nil {
		return err
	}

	agent, err := k8s.GetAgent(ctx, client, bcloudServer)
	if err != nil {
		return err
	}

	if agent == nil {
		// TODO: handle case where agent is not present
		// TODO: detect browser available
		fmt.Fprintf(os.Stderr,
			"Opening linkerd-buoyant agent setup at:\n%s/connect-cluster?linkerd-buoyant=ABC123\n",
			bcloudServer,
		)
	} else {
		// TODO: handle 400s/500s
		resp, err := http.Get(agent.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "%s\n", body)

		fmt.Fprintf(os.Stderr,
			"linkerd-buoyant agent '%s' (%s) found on cluster '%s'.\n",
			agent.Name, agent.Version, currentContext,
		)
		fmt.Fprintf(os.Stderr, "Upgrading to v0.0.28...\n\n") // TODO: retrieve for latest version string from buoyant.cloud

		fmt.Fprintf(os.Stderr, "Agent manifest available at:\n%s\n", agent.URL)
	}

	return nil
}

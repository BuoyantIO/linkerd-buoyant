package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/cli/pkg/k8s"
	"github.com/spf13/cobra"
)

type installCfg struct {
	*config
	agentName string
	noTLS     bool
}

func newCmdInstall(cfg *config) *cobra.Command {
	installCfg := &installCfg{config: cfg}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Buoyant Cloud agent manifest for installation",
		Long: `Output Buoyant Cloud agent manifest for installation.

This command provides the Kubernetes configs necessary to install the Buoyant
Cloud agent on your cluster.

If an agent is not already present, this command provides a manifest to apply to
your cluster, which auto-registers the cluster with Buoyant Cloud using the name
that you've specified.

If an agent is already present, this command retrieves an updated manifest from
Buoyant Cloud and outputs it.

Note that this command requires that the BUOYANT_CLOUD_CLIENT_ID and
BUOYANT_CLOUD_CLIENT_SECRET environment variables are set. To retrieve the correct
values for these variables, visit: https://buoyant.cloud/settings?cli=1.`,
		Example: `  # Default install (no agent on cluster).
  linkerd buoyant install --agent-name=my-new-agent | kubectl apply -f -

  # Obtain a manifest for an agent that already exists on your cluster
  linkerd buoyant install | kubectl apply -f -

  # Install onto a specific cluster
  linkerd buoyant --context test-cluster install | kubectl --context test-cluster apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.New(installCfg.kubeconfig, cfg.kubecontext)
			if err != nil {
				return err
			}

			// get clientID and clientSecret from either k8s or env vars
			var clientID, clientSecret string

			secret, err := client.Secret(cmd.Context())
			if err == nil && secret != nil {
				clientID = string(secret.Data["client_id"])
				clientSecret = string(secret.Data["client_secret"])
			}

			// prefer env vars
			if os.Getenv("BUOYANT_CLOUD_CLIENT_ID") != "" {
				clientID = os.Getenv("BUOYANT_CLOUD_CLIENT_ID")
			}
			if os.Getenv("BUOYANT_CLOUD_CLIENT_SECRET") != "" {
				clientSecret = os.Getenv("BUOYANT_CLOUD_CLIENT_SECRET")
			}

			if clientID == "" || clientSecret == "" {
				fmt.Fprint(
					cfg.stderr,
					"This command requires that the BUOYANT_CLOUD_CLIENT_ID and BUOYANT_CLOUD_CLIENT_SECRET\n",
					"environment variables be set. To retrieve the correct values for these\n",
					"variables, visit: https://buoyant.cloud/settings?cli=1.\n",
				)
				return fmt.Errorf("no credentials set")
			}

			apiClient := bcloudapi.New(clientID, clientSecret, installCfg.bcloudAPI, installCfg.noTLS)

			return install(cmd.Context(), installCfg, client, apiClient)
		},
	}

	cmd.Flags().StringVar(&installCfg.agentName, "agent-name", "", "The name of the agent.")
	cmd.Flags().BoolVar(&installCfg.noTLS, "no-tls", false, "Disable TLS in development mode.")

	cmd.Flags().MarkHidden("insecure")
	return cmd
}

func install(ctx context.Context, cfg *installCfg, client k8s.Client, apiClient bcloudapi.Client) error {
	var identifier bcloudapi.AgentIdentifier
	if cfg.agentName != "" {
		// we need a manifest for a specific name
		identifier = bcloudapi.AgentName(cfg.agentName)
	} else {
		// we need to look at the cluster
		agent, err := client.Agent(ctx)
		if err != nil {
			return err
		}

		if agent.Id != "" {
			identifier = bcloudapi.AgentID(agent.Id)
		} else if agent.Name != "" {
			identifier = bcloudapi.AgentName(agent.Name)
		} else {
			fmt.Fprintf(cfg.stderr,
				"Could not find valid agent installation on cluster. To install agent run:\n%s\n",
				"linkerd buoyant install --agent-name=my-new-agent | kubectl apply -f -",
			)
			return fmt.Errorf("could not find valid agent installation")
		}
	}

	manifest, err := apiClient.GetAgentManifest(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to retrieve agent manifest from bcloud server for agent identifier %T %s: %w", identifier, identifier.Value(), err)
	}

	// output the YAML manifest, this is the only thing that outputs to stdout
	fmt.Fprintf(cfg.stdout, "%s\n", manifest)

	fmt.Fprint(cfg.stderr, "Need help? Message us in the #buoyant-cloud Slack channel:\nhttps://linkerd.slack.com/archives/C01QSTM20BY\n\n")

	return nil
}

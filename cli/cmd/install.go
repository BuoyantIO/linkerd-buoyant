package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/cli/pkg/k8s"
	"github.com/spf13/cobra"
)

type installCfg struct {
	*config
	agentName string
	insecure  bool
}

func newCmdInstall(cfg *config) *cobra.Command {
	installCfg := &installCfg{config: cfg}

	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Buoyant Cloud agent manifest for installation",
		Long: `Output Buoyant Cloud agent manifest for installation.

This command provides the Kubernetes configs necessary to install the Buoyant
Cloud Agent.

If an agent is not already present on the current cluster, this command
will provide you with a manifest that should perform auto registration of
the agent with the name you have specified.

If an agent is already present, this command retrieves an updated manifest from
Buoyant Cloud and outputs it.

Note that this command required CLIENT_ID and CLIENT_SECRET env vars to be set.
To get a CLIENT_ID and CLIENT_SECRET, log in to https://buoyant.cloud.`,
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
			if os.Getenv("CLIENT_ID") != "" {
				clientID = os.Getenv("CLIENT_ID")
			}
			if os.Getenv("CLIENT_SECRET") != "" {
				clientSecret = os.Getenv("CLIENT_SECRET")
			}

			if clientID == "" {
				return errors.New("install command requires setting CLIENT_ID")
			}
			if clientSecret == "" {
				return errors.New("install command requires setting CLIENT_SECRET")
			}

			apiClient := bcloudapi.New(clientID, clientSecret, installCfg.bcloudAPI, !installCfg.insecure)

			return install(cmd.Context(), installCfg, client, apiClient)
		},
	}

	cmd.Flags().StringVar(&installCfg.agentName, "agent-name", "", "The name of the agent.")
	cmd.Flags().BoolVar(&installCfg.insecure, "insecure", false, "Disable TLS in development mode.")

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

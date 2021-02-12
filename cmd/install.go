package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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
  linkerd buoyant --context test-cluster install | kubectl apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(cmd.Context())
		},
	}

	return cmd
}

func install(ctx context.Context) error {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: kubecontext})

	r, err := clientConfig.RawConfig()
	if err != nil {
		return err
	}
	currentContext := r.CurrentContext

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// TODO: handle case where agent is not present
	bcloudID, err := clientset.
		CoreV1().
		Secrets("buoyant-cloud").
		Get(ctx, "buoyant-cloud-id", metav1.GetOptions{})
	if err != nil {
		return err
	}

	bcloudDeploy, err := clientset.
		AppsV1().
		Deployments("buoyant-cloud").
		Get(ctx, "buoyant-cloud-agent", metav1.GetOptions{})
	if err != nil {
		return err
	}

	u := fmt.Sprintf(
		"%s/agent/buoyant-cloud-k8s-%s-%s-%s.yml",
		bcloudServer, bcloudID.Data["name"], bcloudID.Data["id"], bcloudID.Data["downloadKey"],
	)
	resp, err := http.Get(u)
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
		bcloudID.Data["name"], bcloudDeploy.GetAnnotations()["buoyant.cloud/version"], currentContext,
	)
	fmt.Fprintf(os.Stderr, "Upgrading to v0.0.28...\n\n") // TODO: retrieve for latest version string from buoyant.cloud

	fmt.Fprintf(os.Stderr, "Agent manifest available at:\n%s\n\n", u)

	return nil
}

package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	// Load all the auth plugins for the cloud providers.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	kubecontext  string
	kubeconfig   string
	bcloudServer string
)

// Root returns the root linkerd-buoyant command. All subcommands hang off of
// this.
func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "linkerd-buoyant",
		Short: "Manage the Linkerd Buoyant extension",
		Long: `This command manages the Linkerd Buoyant extension.

It enables operational control over the Buoyant Cloud Agent, providing install,
upgrade, and delete functionality`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.PersistentFlags().StringVar(&kubecontext, "context", "", "The name of the kubeconfig context to use")
	if home := homedir.HomeDir(); home != "" {
		root.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Path to the kubeconfig file to use for CLI requests")
	} else {
		root.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests")
	}
	root.PersistentFlags().StringVar(&bcloudServer, "bcloud-server", "https://buoyant.cloud", "Buoyant Cloud server to retrieve manifests from (for testing)")
	root.PersistentFlags().MarkHidden("bcloud-server")

	root.AddCommand(newCmdInstall())

	return root
}

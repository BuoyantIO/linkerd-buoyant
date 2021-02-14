package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	// Load all the auth plugins for the cloud providers.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	kubecontext  string
	kubeconfig   string
	verbose      bool
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

	defaultKubeConfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeConfig = filepath.Join(home, ".kube", "config")
	}

	root.PersistentFlags().StringVar(&kubecontext, "context", "", "The name of the kubeconfig context to use")
	root.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", defaultKubeConfig, "Path to the kubeconfig file to use for CLI requests")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Turn on debug logging")

	// hidden
	root.PersistentFlags().StringVar(&bcloudServer, "bcloud-server", "https://buoyant.cloud", "Buoyant Cloud server to retrieve manifests from (for testing)")
	root.PersistentFlags().MarkHidden("bcloud-server")

	// hidden and unused, to satisfy linkerd extension interface
	var apiAddr, l5dVersion string
	root.PersistentFlags().StringVar(&apiAddr, "api-addr", "", "")
	root.PersistentFlags().StringVarP(&l5dVersion, "linkerd-namespace", "l", "", "")
	root.PersistentFlags().MarkHidden("linkerd-namespace")
	root.PersistentFlags().MarkHidden("api-addr")

	root.AddCommand(newCmdInstall())
	root.AddCommand(newCmdUninstall())

	return root
}

func printVerbosef(format string, a ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
	}
}

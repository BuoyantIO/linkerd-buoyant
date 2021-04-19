package cmd

import (
	"os"
	"path/filepath"

	"github.com/buoyantio/linkerd-buoyant/pkg/version"
	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// Root returns the root linkerd-buoyant command. All subcommands hang off of
// this.
func Root() *cobra.Command {
	cfg := &config{
		// special handling for Windows, on all other platforms these resolve to
		// os.Stdout and os.Stderr, thanks to https://github.com/mattn/go-colorable
		stdout: color.Output,
		stderr: color.Error,
	}

	root := &cobra.Command{
		Use:   version.LinkerdBuoyant,
		Short: "Manage the Linkerd Buoyant extension",
		Long: `linkerd-buoyant manages the Buoyant Cloud agent.

It enables operational control over the Buoyant Cloud agent, providing install,
upgrade, and delete functionality.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	defaultKubeConfig := ""
	if home := homedir.HomeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		_, err := os.Stat(kubeconfig)
		if !os.IsNotExist(err) {
			defaultKubeConfig = kubeconfig
		}
	}

	// global flags
	root.PersistentFlags().StringVar(&cfg.kubecontext, "context", "", "The name of the kubeconfig context to use")
	root.PersistentFlags().StringVar(&cfg.kubeconfig, "kubeconfig", defaultKubeConfig, "Path to the kubeconfig file to use for CLI requests")
	root.PersistentFlags().BoolVarP(&cfg.verbose, "verbose", "v", false, "Turn on debug logging")

	// hidden flags
	root.PersistentFlags().StringVar(&cfg.bcloudServer, "bcloud-server", "https://buoyant.cloud", "Buoyant Cloud server to retrieve manifests from (for testing)")
	root.PersistentFlags().MarkHidden("bcloud-server")

	// hidden and unused flags, to satisfy linkerd extension interface
	var apiAddr, l5dVersion string
	root.PersistentFlags().StringVar(&apiAddr, "api-addr", "", "")
	root.PersistentFlags().StringVarP(&l5dVersion, "linkerd-namespace", "l", "", "")
	root.PersistentFlags().MarkHidden("linkerd-namespace")
	root.PersistentFlags().MarkHidden("api-addr")

	// add all subcommands
	root.AddCommand(newCmdCheck(cfg))
	root.AddCommand(newCmdInstall(cfg))
	root.AddCommand(newCmdUninstall(cfg))
	root.AddCommand(newCmdVersion(cfg))
	root.AddCommand(newCmdDashboard(cfg, browser.OpenURL))

	return root
}

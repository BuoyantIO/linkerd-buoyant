package cmd

import "github.com/spf13/cobra"

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

	root.AddCommand(install())

	return root
}

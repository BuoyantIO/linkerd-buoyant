package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCmdDashboard(cfg *config, openURL openURL) *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard [flags]",
		Args:  cobra.NoArgs,
		Short: "Open the Buoyant Cloud dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cfg.stdout, "Opening Buoyant Cloud dashboard in the default browser")

			err := openURL(cfg.bcloudServer)
			if err != nil {
				fmt.Fprintln(cfg.stderr, "Failed to open dashboard automatically")
				fmt.Fprintf(cfg.stderr, "Visit %s in your browser to view the dashboard\n", cfg.bcloudServer)
			}

			return nil
		},
	}
}

package cmd

import (
	"github.com/spf13/cobra"
)

func newCmdDashboard(cfg *config, openURL openURL) *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard [flags]",
		Args:  cobra.NoArgs,
		Short: "Open the Buoyant Cloud dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return openURL(cfg.bcloudServer)
		},
	}
}

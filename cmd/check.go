package cmd

import (
	"fmt"
	"os"
	"time"

	pkghealthcheck "github.com/buoyantio/linkerd-buoyant/pkg/healthcheck"
	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/linkerd/linkerd2/pkg/healthcheck"
	"github.com/spf13/cobra"
)

type checkConfig struct {
	config
	output string
	wait   time.Duration
}

func newCmdCheck(cfg config) *cobra.Command {
	checkCfg := checkConfig{config: cfg}

	cmd := &cobra.Command{
		Use:   "check [flags]",
		Args:  cobra.NoArgs,
		Short: "Check the Buoyant Cloud Agent installation for potential problems",
		Long: `Check the Buoyant Cloud Agent installation for potential problems.

The check command will perform a series of checks to validate that the
linkerd-buoyant CLI and Buoyant Cloud Agent are configured correctly. If the
command encounters a failure it will print additional information about the
failure and exit with a non-zero exit code.`,
		Example: `  # Default check.
  linkerd-buoyant check
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.New(checkCfg.kubeconfig, checkCfg.kubecontext, checkCfg.bcloudServer)
			if err != nil {
				return err
			}

			return check(checkCfg, client)
		},
	}

	cmd.Flags().StringVarP(&checkCfg.output, "output", "o", healthcheck.TableOutput, "Output format. One of: table, json")
	cmd.Flags().DurationVar(&checkCfg.wait, "wait", 1*time.Minute, "Maximum allowed time for all tests to pass")

	// hidden and unused, to satisfy linkerd extension interface
	var proxy bool
	var namespace string
	cmd.Flags().BoolVar(&proxy, "proxy", false, "")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "")
	cmd.Flags().MarkHidden("proxy")
	cmd.Flags().MarkHidden("namespace")

	return cmd
}

func check(cfg checkConfig, client k8s.Client) error {
	if cfg.output != healthcheck.TableOutput && cfg.output != healthcheck.JSONOutput {
		return fmt.Errorf(
			"Invalid output type '%s'. Supported output types are: %s, %s",
			cfg.output, healthcheck.TableOutput, healthcheck.JSONOutput,
		)
	}

	hc := pkghealthcheck.NewHealthChecker(
		client,
		&healthcheck.Options{
			RetryDeadline: time.Now().Add(cfg.wait),
		},
	)

	hc.AppendCategories(hc.L5dBuoyantCategory())
	success := healthcheck.RunChecks(cfg.stdout, cfg.stderr, hc, cfg.output)

	if !success {
		os.Exit(1)
	}

	return nil
}

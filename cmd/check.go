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

var (
	checkOutput string
	checkWait   time.Duration
)

func newCmdCheck() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Kubernetes resources to check the linkerd-buoyant extension",
		Long: `Output Kubernetes resources to check the linkerd-buoyant extension.

This command provides all Kubernetes namespace-scoped and cluster-scoped
resources (e.g services, deployments, RBACs, etc.) necessary to uninstall the
linkerd-buoyant extension.`,
		Example: `  # Default uninstall.
  linkerd buoyant uninstall | kubectl delete -f -

  # Unnstall from a specific cluster
  linkerd buoyant --context test-cluster uninstall | kubectl delete --context test-cluster -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return check()
		},
	}

	cmd.Flags().StringVarP(&checkOutput, "output", "o", healthcheck.TableOutput, "Output format. One of: table, json")
	cmd.Flags().DurationVar(&checkWait, "wait", 10*time.Second, "Maximum allowed time for all tests to pass")

	// hidden and unused, to satisfy linkerd extension interface
	var proxy bool
	var namespace string
	cmd.Flags().BoolVar(&proxy, "proxy", false, "")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "")
	cmd.Flags().MarkHidden("proxy")
	cmd.Flags().MarkHidden("namespace")

	return cmd
}

func check() error {
	if checkOutput != healthcheck.TableOutput && checkOutput != healthcheck.JSONOutput {
		return fmt.Errorf(
			"Invalid output type '%s'. Supported output types are: %s, %s",
			checkOutput, healthcheck.TableOutput, healthcheck.JSONOutput,
		)
	}

	client, err := k8s.New(kubeconfig, kubecontext)
	if err != nil {
		return err
	}

	hc := pkghealthcheck.NewHealthChecker(
		client,
		&healthcheck.Options{
			RetryDeadline: time.Now().Add(checkWait),
		},
	)

	hc.AppendCategories(hc.L5dBuoyantCategory())
	success := healthcheck.RunChecks(stdout, stderr, hc, checkOutput)

	if !success {
		os.Exit(1)
	}

	return nil
}

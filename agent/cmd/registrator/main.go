package registrator

import (
	"context"
	"flag"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/flags"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/registrator"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

func dieIf(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Main executes the registrator subcommand
func Main(args []string) {
	cmd := flag.NewFlagSet("registrator", flag.ExitOnError)

	apiAddr := cmd.String("api-addr", "api.buoyant.cloud:443", "address of the Buoyant Cloud API")
	kubeConfigPath := cmd.String("kubeconfig", "", "path to kube config")
	insecure := cmd.Bool("insecure", false, "disable TLS in development mode")
	agentMetadataMap := cmd.String("agent-metadata-map", "agent-metadata", "the name of the agent metadata map")

	clientID, clientSecret := flags.ConfigureAndParseAgentParams(cmd, args)

	// setup kubernetes client
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	if *kubeConfigPath != "" {
		rules.ExplicitPath = *kubeConfigPath
	}

	k8sConfig, err := clientcmd.
		NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).
		ClientConfig()
	dieIf(err)

	k8sAPI, err := l5dk8s.NewAPIForConfig(k8sConfig, "", nil, 0)
	dieIf(err)

	// perform agent registration

	secure := !*insecure
	bcloudApiClient := bcloudapi.New(clientID, clientSecret, *apiAddr, secure)
	agentRegistrator := registrator.New(bcloudApiClient, k8sAPI, *agentMetadataMap)

	agentInfo, err := agentRegistrator.EnsureRegistered(context.Background())
	dieIf(err)
	log.Infof("Obtained agent info: %+v", agentInfo)
}

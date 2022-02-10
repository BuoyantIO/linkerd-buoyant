package registrator

import (
	"context"
	"flag"
	"os"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/registrator"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	// Load all the auth plugins for the cloud providers.
	// This enables connecting to a k8s cluster from outside the cluster, during
	// development.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func dieIf(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Main executes the registrator subcommand
func Main(args []string) {
	cmd := flag.NewFlagSet("registrator", flag.ExitOnError)

	clientID := cmd.String("client-id", "", "bcloud client id, takes precedence over CLIENT_ID env var")
	clientSecret := cmd.String("client-secret", "", "bcloud client secret, takes precedence over CLIENT_SECRET env var")
	apiAddr := cmd.String("api-addr", "api.buoyant.cloud:443", "address of the Buoyant Cloud API")
	kubeConfigPath := cmd.String("kubeconfig", "", "path to kube config")
	logLevel := cmd.String("log-level", "info", "log level, must be one of: panic, fatal, error, warn, info, debug, trace")
	insecure := cmd.Bool("insecure", false, "disable TLS in development mode")

	// klog flags
	klog.InitFlags(nil)
	flag.Set("stderrthreshold", "FATAL")
	flag.Set("logtostderr", "true")
	flag.Set("v", "0")

	cmd.Parse(args)

	// set global log level and format
	level, err := log.ParseLevel(*logLevel)
	dieIf(err)
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	// klog flags for debugging
	if level >= log.DebugLevel {
		flag.Set("stderrthreshold", "INFO")
		flag.Set("v", "6") // At 7 and higher, authorization tokens get logged.
	}
	// pipe klog entries to logrus
	klog.SetOutput(log.StandardLogger().Writer())

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
	id := os.Getenv("CLIENT_ID")
	if *clientID != "" {
		id = *clientID
	}
	if id == "" {
		log.Fatal("missing client id! set -client-id flag or CLIENT_ID env var")
	}
	log.Debugf("using bcloud client id %s", id)

	secret := os.Getenv("CLIENT_SECRET")
	if *clientSecret != "" {
		secret = *clientSecret
	}
	if secret == "" {
		log.Fatal("missing bcloud client secret! set -client-secret flag or CLIENT_SECRET env var")
	}

	secure := !*insecure
	bcloudApiClient := bcloudapi.New(id, secret, *apiAddr, secure)
	agentRegistrator := registrator.New(bcloudApiClient, k8sAPI)

	agentInfo, err := agentRegistrator.EnsureRegistered(context.Background())
	dieIf(err)
	log.Infof("Obtained agent info: %+v", agentInfo)
}

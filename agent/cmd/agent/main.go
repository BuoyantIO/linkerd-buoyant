package agent

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/bcloudapi"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/flags"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/handler"
	"github.com/buoyantio/linkerd-buoyant/agent/pkg/k8s"
	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	l5dApi "github.com/linkerd/linkerd2/controller/gen/client/clientset/versioned"
	"github.com/linkerd/linkerd2/pkg/admin"
	l5dk8s "github.com/linkerd/linkerd2/pkg/k8s"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/clientcmd"
)

func dieIf(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Main executes the agent subcommand
func Main(args []string) {
	cmd := flag.NewFlagSet("agent", flag.ExitOnError)

	apiAddr := cmd.String("api-addr", "api.buoyant.cloud:443", "address of the Buoyant Cloud API")
	adminAddr := cmd.String("admin-addr", ":9990", "address of agent admin server")
	grpcAddr := cmd.String("grpc-addr", "api.buoyant.cloud:443", "address of the Buoyant Cloud gRPC API")
	kubeConfigPath := cmd.String("kubeconfig", "", "path to kube config")
	localMode := cmd.Bool("local-mode", false, "enable port forwarding for local development")
	insecure := cmd.Bool("insecure", false, "disable TLS in development mode")
	agentID := cmd.String("agent-id", "", "the ID of the agent")

	clientID, clientSecret := flags.ConfigureAndParseAgentParams(cmd, args)
	if agentID == nil || *agentID == "" {
		log.Fatal("missing agent id! set -agent-id flag")
	}

	// handle interrupts
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	shutdown := make(chan struct{}, 1)

	// setup kubernetes clients and shared informers
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

	l5dClient, err := l5dApi.NewForConfig(k8sConfig)
	dieIf(err)

	sharedInformers := informers.NewSharedInformerFactory(k8sAPI.Interface, 10*time.Minute)

	k8sClient := k8s.NewClient(sharedInformers, k8sAPI, l5dClient, *localMode)

	// wait for discovery API to load

	log.Info("waiting for Kubernetes API availability")
	populateGroupList := func() (done bool, err error) {
		_, err = k8sAPI.Discovery().ServerGroups()
		if err != nil {
			log.Debug("cannot reach Kubernetes API; retrying")
			return false, nil
		}
		log.Info("Kubernetes API reached")
		return true, nil
	}
	err = wait.PollImmediate(time.Second, time.Minute, populateGroupList)
	dieIf(err)

	// create bcloud grpc api client and streams

	secure := !*insecure
	bcloudApiClient := bcloudapi.New(clientID, clientSecret, *apiAddr, secure)
	perRPCCreds := bcloudApiClient.Credentials(context.Background(), *agentID)
	opts := []grpc.DialOption{grpc.WithPerRPCCredentials(perRPCCreds)}
	if *insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	conn, err := grpc.Dial(*grpcAddr, opts...)
	dieIf(err)

	bcloudClient := pb.NewApiClient(conn)
	apiClient := api.NewClient(bcloudClient)

	// create handlers
	eventHandler := handler.NewEvent(k8sClient, apiClient)
	workloadHandler := handler.NewWorkload(k8sClient, apiClient)

	linkerdInfoHandler := handler.NewLinkerdInfo(k8sClient, apiClient)
	manageAgentHandler := handler.NewManageAgent(k8sClient, apiClient)

	// start shared informer and wait for sync
	err = k8sClient.Sync(shutdown, 60*time.Second)
	dieIf(err)

	// start api client stream management logic
	go apiClient.Start()

	// start handlers
	go eventHandler.Start(sharedInformers)
	go workloadHandler.Start(sharedInformers)
	go linkerdInfoHandler.Start()
	go manageAgentHandler.Start()

	// run admin server
	adminServer := admin.NewServer(*adminAddr)
	go adminServer.ListenAndServe()

	// wait for shutdown
	<-stop
	log.Info("shutting down")
	workloadHandler.Stop()
	linkerdInfoHandler.Stop()
	manageAgentHandler.Stop()
	close(shutdown)
}

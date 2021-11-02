package main

import (
	"crypto/tls"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buoyantio/linkerd-buoyant/agent/pkg/api"
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

func main() {
	bcloudID := flag.String("bcloud-id", "", "bcloud id, takes precedence over BUOYANT_CLOUD_ID env var")
	bcloudKey := flag.String("bcloud-key", "", "bcloud key, takes precedence over BUOYANT_CLOUD_KEY env var")
	adminAddr := flag.String("admin-addr", ":9990", "address of agent admin server")
	grpcAddr := flag.String("grpc-addr", "api.buoyant.cloud:443", "address of the Buoyant Cloud API")
	kubeConfigPath := flag.String("kubeconfig", "", "path to kube config")
	logLevel := flag.String("log-level", "info", "log level, must be one of: panic, fatal, error, warn, info, debug, trace")
	localMode := flag.Bool("local-mode", false, "enable port forwarding for local development")
	insecure := flag.Bool("insecure", false, "disable TLS in development mode")

	// klog flags
	klog.InitFlags(nil)
	flag.Set("stderrthreshold", "FATAL")
	flag.Set("logtostderr", "true")
	flag.Set("v", "0")

	flag.Parse()

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

	// handle interrupts

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	shutdown := make(chan struct{}, 1)

	// read the bcloud ID from flag or environment variable

	id := os.Getenv("BUOYANT_CLOUD_ID")
	if *bcloudID != "" {
		id = *bcloudID
	}
	if id == "" {
		log.Fatal("missing bcloud id! set -bcloud-id flag or BUOYANT_CLOUD_ID env var")
	}
	log.Debugf("using bcloud id %s", id)

	key := os.Getenv("BUOYANT_CLOUD_KEY")
	if *bcloudKey != "" {
		key = *bcloudKey
	}
	if key == "" {
		log.Fatal("missing bcloud key! set -bcloud-key flag or BUOYANT_CLOUD_KEY env var")
	}

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

	var opts grpc.DialOption
	if *insecure {
		opts = grpc.WithInsecure()
	} else {
		creds := credentials.NewTLS(&tls.Config{})
		opts = grpc.WithTransportCredentials(creds)
	}
	conn, err := grpc.Dial(*grpcAddr, opts)
	dieIf(err)

	bcloudClient := pb.NewApiClient(conn)
	apiClient := api.NewClient(id, key, bcloudClient)

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

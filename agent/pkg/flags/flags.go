package flags

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

// ConfigureAndParse adds flags that are common to all agent go processes. This
// func calls flag.Parse(), so it should be called after all other flags have
// been configured. On success it returns the clientID and clientSecret credentials
// the are needed for the agent to communicate with the Bcloud API
func ConfigureAndParseAgentParams(cmd *flag.FlagSet, args []string) (string, string) {
	// klog flags
	klog.InitFlags(nil)
	flag.Set("stderrthreshold", "FATAL")
	flag.Set("logtostderr", "true")
	flag.Set("v", "0")

	logLevel := cmd.String("log-level", log.InfoLevel.String(),
		"log level, must be one of: panic, fatal, error, warn, info, debug")
	clientID := cmd.String("client-id", "", "bcloud client id, takes precedence over BUOYANT_CLOUD_CLIENT_ID env var")
	clientSecret := cmd.String("client-secret", "", "bcloud client secret, takes precedence over BUOYANT_CLOUD_CLIENT_SECRET env var")

	cmd.Parse(args)
	setLogLevel(*logLevel)
	return credentials(clientID, clientSecret)
}

// credentials ensures that client id and client secret credentials are
// provided in either via command line parameters or env varibles, giving
// preference to the former.
func credentials(clientID, clientSecret *string) (string, string) {
	id := os.Getenv("BUOYANT_CLOUD_CLIENT_ID")
	if *clientID != "" {
		id = *clientID
	}
	if id == "" {
		log.Fatal("missing client id! set -client-id flag or BUOYANT_CLOUD_CLIENT_ID env var")
	}
	log.Debugf("using bcloud client id %s", id)

	secret := os.Getenv("BUOYANT_CLOUD_CLIENT_SECRET")
	if *clientSecret != "" {
		secret = *clientSecret
	}
	if secret == "" {
		log.Fatal("missing bcloud client secret! set -client-secret flag or BUOYANT_CLOUD_CLIENT_SECRET env var")
	}
	return id, secret
}

func setLogLevel(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("invalid log-level: %s", logLevel)
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	// klog flags for debugging
	if level >= log.DebugLevel {
		flag.Set("stderrthreshold", "INFO")
		flag.Set("v", "6") // At 7 and higher, authorization tokens get logged.
	}
	// pipe klog entries to logrus
	klog.SetOutput(log.StandardLogger().Writer())
}

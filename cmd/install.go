package cmd

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// openURL allows mocking the browser.OpenURL function, so our tests do not open
// a browser window.
type openURL func(url string) error

func newCmdInstall(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [flags]",
		Args:  cobra.NoArgs,
		Short: "Output Buoyant Cloud agent manifest for installation",
		Long: `Output Buoyant Cloud agent manifest for installation.

This command provides the Kubernetes configs necessary to install the Buoyant
Cloud Agent.

If an agent is not already present on the current cluster, this command
redirects the user to Buoyant Cloud to set up a new agent. Once the new agent is
set up, this command will output the agent manifest.

If an agent is already present, this command retrieves an updated manifest from
Buoyant Cloud and outputs it.`,
		Example: `  # Default install.
  linkerd buoyant install | kubectl apply -f -

  # Install onto a specific cluster
  linkerd buoyant --context test-cluster install | kubectl --context test-cluster apply -f -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := k8s.New(cfg.kubeconfig, cfg.kubecontext, cfg.bcloudServer)
			if err != nil {
				return err
			}

			return install(cmd.Context(), cfg, client, browser.OpenURL)
		},
	}

	return cmd
}

func install(ctx context.Context, cfg *config, client k8s.Client, openURL openURL) error {
	agent, err := client.Agent(ctx)
	if err != nil {
		return err
	}

	var agentURL string
	if agent != nil {
		// existing agent on cluster
		agentURL = agent.URL

		cfg.printVerbosef("Agent found on cluster, latest manifest URL:\n%s", agentURL)
	} else {
		// new agent
		agentURL, err = newAgentURL(cfg, openURL)
		if err != nil {
			return err
		}

		cfg.printVerbosef("No agent found on cluster, new manifest URL:\n%s", agentURL)
	}

	resp, err := http.Get(agentURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(cfg.stderr,
			"Unexpected HTTP status code %d for URL:\n%s\n",
			resp.StatusCode, agentURL,
		)
		return fmt.Errorf("failed to retrieve agent manifest from %s", agentURL)
	}

	if resp.Header.Get("Content-type") != "text/yaml" {
		return fmt.Errorf("unexpected Content-Type '%s' from %s", resp.Header.Get("Content-type"), agentURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// output the YAML manifest, this is the only thing that outputs to stdout
	fmt.Fprintf(cfg.stdout, "%s\n", body)

	fmt.Fprintf(cfg.stderr, "Agent manifest available at:\n%s\n\n", agentURL)

	return nil
}

func newAgentURL(cfg *config, openURL openURL) (string, error) {
	agentUID := genUniqueID()

	connectURL := fmt.Sprintf("%s/connect-cluster?linkerd-buoyant=%s", cfg.bcloudServer, agentUID)
	err := openURL(connectURL)
	if err == nil {
		fmt.Fprintf(cfg.stderr, "Opening Buoyant Cloud agent setup at:\n%s\n", connectURL)
	} else {
		fmt.Fprintf(cfg.stderr, "Visit this URL to set up the Buoyant Cloud agent:\n%s\n\n", connectURL)
	}

	// start polling
	fmt.Fprintf(cfg.stderr, "Waiting for agent setup completion...\n")

	// don't automatically follow redirect, we want to capture the manifest URL
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	connectAgentURL := fmt.Sprintf("%s/connect-agent?linkerd-buoyant=%s", cfg.bcloudServer, agentUID)
	cfg.printVerbosef("Polling: %s", connectAgentURL)

	// only exit on 3 consecutive failures
	retries := 0
	for {
		resp, err := client.Get(connectAgentURL)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusAccepted {
			// still polling
			retries = 0
			time.Sleep(time.Second)
			continue
		}

		if resp.StatusCode != http.StatusPermanentRedirect {
			retries++
			if retries < 3 {
				continue
			}

			return "", fmt.Errorf("setup failed, unexpected HTTP status code %d for URL %s", resp.StatusCode, connectAgentURL)
		}

		// successful 308, get the agent YAML URL
		url, err := resp.Location()
		if err != nil {
			return "", err
		}

		cfg.printVerbosef("Agent setup completed, redirecting to: %s", url.String())

		return url.String(), nil
	}
}

// genUniqueID makes a random 16 character ascii string.
func genUniqueID() string {
	timeBytes := new([8]byte)[0:8]
	binary.BigEndian.PutUint64(timeBytes, uint64(time.Now().UnixNano()))

	randBytes := new([8]byte)[0:8]
	binary.BigEndian.PutUint64(randBytes, uint64(rand.Int63()))

	hasher := sha1.New()
	hasher.Write([]byte("linkerd x buoyant == <3"))
	hasher.Write(timeBytes)
	hasher.Write(randBytes)

	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))[0:16]
}

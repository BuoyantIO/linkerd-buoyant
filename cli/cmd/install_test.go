package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/cli/pkg/k8s"
	"github.com/buoyantio/linkerd-buoyant/cli/pkg/version"
)

const connectAgentPath = "/connect-agent"

func TestInstallNewAgent(t *testing.T) {
	totalRequests := 0
	connectRequests := 0
	redirectRequests := 0
	agentUID := "'"
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			totalRequests++
			switch r.URL.Path {
			case connectAgentPath:
				connectRequests++
				agentUID = r.URL.Query().Get(version.LinkerdBuoyant)
				if connectRequests == 1 {
					w.WriteHeader(http.StatusAccepted)
					return
				}

				if connectRequests == 2 {
					w.WriteHeader(http.StatusBadGateway)
					return
				}

				http.Redirect(w, r, "/agent-yaml-redirect", http.StatusPermanentRedirect)
			case "/agent-yaml-redirect":
				redirectRequests++
				w.Header().Set("Content-Type", "text/yaml")
				w.Write([]byte("fake-yaml"))
			}
		},
	))
	defer ts.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: ts.URL,
	}

	client := &k8s.MockClient{}
	err := install(context.TODO(), cfg, client, mockOpenURL)
	if err != nil {
		t.Error(err)
	}

	if stdout.String() != "fake-yaml\n" {
		t.Errorf("Expected: [fake-yaml], Got: [%s]", stdout.String())
	}
	expBrowserURL := fmt.Sprintf("%s/connect-cluster?linkerd-buoyant=%s", ts.URL, agentUID)
	if !strings.Contains(stderr.String(), expBrowserURL) {
		t.Errorf("Expected stderr to contain [%s], Got: [%s]", expBrowserURL, stderr.String())
	}
	if totalRequests != 4 {
		t.Errorf("Expected 4 total requests, called %d times", totalRequests)
	}
	if connectRequests != 3 {
		t.Errorf("Expected 3 /connect-agent requests, called %d times", connectRequests)
	}
	if redirectRequests != 1 {
		t.Errorf("Expected 1 /agent-yaml-redirect request, called %d times", redirectRequests)
	}
}

func TestInstalWithPollingFailures(t *testing.T) {
	totalRequests := 0
	connectRequests := 0
	agentUID := "'"
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			totalRequests++
			switch r.URL.Path {
			case connectAgentPath:
				agentUID = r.URL.Query().Get(version.LinkerdBuoyant)
				connectRequests++
				w.WriteHeader(http.StatusBadGateway)
			}
		},
	))
	defer ts.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: ts.URL,
	}

	client := &k8s.MockClient{}
	err := install(context.TODO(), cfg, client, mockOpenURL)
	expErr := fmt.Errorf("setup failed, unexpected HTTP status code 502 for URL %s/connect-agent?linkerd-buoyant=%s", ts.URL, agentUID)
	if !reflect.DeepEqual(err, expErr) {
		t.Errorf("Expected error: %s, Got: %s", expErr, err)
	}

	if totalRequests != maxPollingRetries {
		t.Errorf("Expected %d total requests, called %d times", maxPollingRetries, totalRequests)
	}
	if connectRequests != maxPollingRetries {
		t.Errorf("Expected %d /connect-agent requests, called %d times", maxPollingRetries, connectRequests)
	}
}

func TestInstallWithPollingApiErrors(t *testing.T) {
	totalRequests := 0
	connectRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			totalRequests++
			switch r.URL.Path {
			case connectAgentPath:
				connectRequests++
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				rsp, _ := json.Marshal(jsonError{Error: "fatal API error"})
				w.Write(rsp)
			}
		},
	))
	defer ts.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: ts.URL,
	}

	client := &k8s.MockClient{}
	err := install(context.TODO(), cfg, client, mockOpenURL)
	expErr := errors.New("setup failed, fatal API error")
	if !reflect.DeepEqual(err, expErr) {
		t.Errorf("Expected error: %s, Got: %s", expErr, err)
	}

	if totalRequests != maxPollingRetries {
		t.Errorf("Expected %d total requests, called %d times", maxPollingRetries, totalRequests)
	}
	if connectRequests != maxPollingRetries {
		t.Errorf("Expected %d /connect-agent requests, called %d times", maxPollingRetries, connectRequests)
	}
}

func TestInstallExistingAgent(t *testing.T) {
	totalRequests := 0
	connectRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			totalRequests++
			if r.URL.Path == "/connect-agent-url" {
				connectRequests++
				w.Header().Set("Content-Type", "text/yaml")
				w.Write([]byte("fake-yaml"))
			}
		},
	))
	defer ts.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: ts.URL,
	}

	client := &k8s.MockClient{
		MockAgent: &k8s.Agent{
			Name:    "name",
			URL:     ts.URL + "/connect-agent-url",
			Version: "version",
		},
	}
	err := install(context.TODO(), cfg, client, mockOpenURL)
	if err != nil {
		t.Error(err)
	}

	if stdout.String() != "fake-yaml\n" {
		t.Errorf("Expected: [fake-yaml], Got: [%s]", stdout.String())
	}
	if totalRequests != 1 {
		t.Errorf("Expected 1 total request, called %d times", totalRequests)
	}
	if connectRequests != 1 {
		t.Errorf("Expected 1 /connect-agent-url request, called %d times", connectRequests)
	}
}

func TestInstallBadStatus(t *testing.T) {
	totalRequests := 0
	connectRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			totalRequests++
			if r.URL.Path == "/connect-agent-url" {
				connectRequests++
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	))
	defer ts.Close()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: ts.URL,
	}

	client := &k8s.MockClient{
		MockAgent: &k8s.Agent{
			Name:    "name",
			URL:     ts.URL + "/connect-agent-url",
			Version: "version",
		},
	}
	expErr := fmt.Errorf("failed to retrieve agent manifest from %s", client.MockAgent.URL)
	err := install(context.TODO(), cfg, client, mockOpenURL)
	if !reflect.DeepEqual(err, expErr) {
		t.Errorf("Expected error: %s, Got: %s", expErr, err)
	}

	if stdout.String() != "" {
		t.Errorf("Expected: no stdout, Got: [%s]", stdout.String())
	}
	if totalRequests != 1 {
		t.Errorf("Expected 1 total request, called %d times", totalRequests)
	}
	if connectRequests != 1 {
		t.Errorf("Expected 1 /connect-agent-url request, called %d times", connectRequests)
	}
}

func mockOpenURL(string) error {
	return nil
}

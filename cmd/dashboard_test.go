package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
)

func TestDashboardNoAgent(t *testing.T) {
	stdout := &bytes.Buffer{}

	cfg := &config{
		stdout:       stdout,
		bcloudServer: "http://example.com",
	}

	client := &k8s.MockClient{
		MockAgent: nil,
	}
	err := dashboardCmd(context.TODO(), cfg, client, mockOpenURL)
	if err != nil {
		t.Error(err)
	}

	expStdout := "Agent not installed, run linkerd-buoyant install\n"
	if stdout.String() != expStdout {
		t.Errorf("Expected %s, Got %s", expStdout, stdout.String())
	}
}

func TestDashboardAgentPresent(t *testing.T) {
	cfg := &config{bcloudServer: "http://example.com"}

	client := &k8s.MockClient{
		MockAgent: &k8s.Agent{},
	}

	var openedUrl string
	mockOpen := func(url string) error {
		openedUrl = url
		return nil
	}

	err := dashboardCmd(context.TODO(), cfg, client, mockOpen)
	if err != nil {
		t.Error(err)
	}

	if openedUrl != cfg.bcloudServer {
		t.Errorf("Expected to open url %s, Got %s", cfg.bcloudServer, openedUrl)
	}
}

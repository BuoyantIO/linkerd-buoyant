package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestDashboard(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: "http://example.com",
	}

	var openedUrl string
	mockOpen := func(url string) error {
		openedUrl = url
		return nil
	}

	cmd := newCmdDashboard(cfg, mockOpen)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Error(err)
	}

	if openedUrl != cfg.bcloudServer {
		t.Fatalf("Expected to open url %s, Got %s", cfg.bcloudServer, openedUrl)
	}
}

func TestDashboardFailure(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := &config{
		stdout:       stdout,
		stderr:       stderr,
		bcloudServer: "http://example.com",
	}

	mockOpen := func(string) error {
		return errors.New("browser failuer")
	}

	cmd := newCmdDashboard(cfg, mockOpen)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Error(err)
	}

	if !strings.Contains(stderr.String(), cfg.bcloudServer) {
		t.Errorf("Expected stderr to contain [%s], Got: [%s]", cfg.bcloudServer, stderr.String())
	}
}

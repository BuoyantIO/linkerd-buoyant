package cmd

import (
	"testing"
)

func TestDashboard(t *testing.T) {
	cfg := &config{bcloudServer: "http://example.com"}

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

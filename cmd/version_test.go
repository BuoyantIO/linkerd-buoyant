package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
)

func TestVersion(t *testing.T) {
	fixtures := []*struct {
		testName string
		cfg      versionConfig
		stdout   string
		stderr   string
	}{
		{
			"defaults",
			versionConfig{},
			"CLI version:   undefined\nAgent version: fake-version\n",
			"",
		},
		{
			"short",
			versionConfig{short: true},
			"undefined\nfake-version\n",
			"",
		},
		{
			"cliOnly",
			versionConfig{cliOnly: true},
			"CLI version:   undefined\n",
			"",
		},
		{
			"short and cliOnly",
			versionConfig{short: true, cliOnly: true},
			"undefined\n",
			"",
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			tc.cfg.stdout = stdout
			tc.cfg.stderr = stderr

			client := &k8s.MockClient{
				MockAgent: &k8s.Agent{Version: "fake-version"},
			}
			err := versionCmd(context.TODO(), tc.cfg, client)
			if err != nil {
				t.Error(err)
			}

			if stdout.String() != tc.stdout {
				t.Errorf("Expected stdout to be [%s], Got [%s]", tc.stdout, stdout.String())
			}
			if stderr.String() != tc.stderr {
				t.Errorf("Expected stderr to be [%s], Got [%s]", tc.stderr, stderr.String())
			}
		})
	}
}

package cmd

import (
	"bytes"
	"testing"
)

func TestPrintVerbosef(t *testing.T) {
	stderr := &bytes.Buffer{}
	cfg := config{
		stderr: stderr,
	}

	cfg.printVerbosef("test output")
	if stderr.String() != "" {
		t.Errorf("Expected not verbose output, got: [%s]", stderr.String())
	}

	cfg.verbose = true
	cfg.printVerbosef("test output")
	if stderr.String() != "test output\n" {
		t.Errorf("Expect verbose output [test output], got: [%s]", stderr.String())
	}
}

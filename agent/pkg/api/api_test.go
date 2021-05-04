package api

import "testing"

func TestNewClient(t *testing.T) {
	c := NewClient("", "", nil)
	if c == nil {
		t.Errorf("Unexpected nil client")
	}
}

package version

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersion(t *testing.T) {
	versionRsp := map[string]string{
		LinkerdBuoyant: "fake-version",
	}

	j, err := json.Marshal(versionRsp)
	if err != nil {
		t.Fatalf("JSON marshal failed with: %s", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write(j)
		}),
	)
	defer ts.Close()

	version, err := Get(context.Background(), ts.Client(), ts.URL)
	if err != nil {
		t.Error(err)
	}
	if version != "fake-version" {
		t.Errorf("Unexpected version: [%s], Expected: [fake-version]", version)
	}
}

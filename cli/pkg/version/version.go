package version

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// LinkerdBuoyant defines the name of the CLI application.
const LinkerdBuoyant = "linkerd-buoyant"

// Version is updated automatically as part of the build process, and is the
// ground source of truth for the current process's build version.
//
// DO NOT EDIT
var Version = "undefined"

// Get retrieves the current linkerd-buoyant version from the Buoyant Cloud
// server.
func Get(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	rsp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected version response: %s", rsp.Status)
	}

	bytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	var versionRsp map[string]string
	err = json.Unmarshal(bytes, &versionRsp)
	if err != nil {
		return "", err
	}

	version, ok := versionRsp[LinkerdBuoyant]
	if !ok {
		return "", fmt.Errorf("unrecognized version response: %s", bytes)
	}

	return version, nil
}

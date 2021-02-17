package cmd

import (
	"errors"
	"reflect"
	"testing"

	"github.com/buoyantio/linkerd-buoyant/pkg/k8s"
)

func TestCheck(t *testing.T) {
	expErr := errors.New("Invalid output type 'bad-output'. Supported output types are: table, json")
	checkCfg := &checkConfig{
		output: "bad-output",
	}

	err := check(checkCfg, &k8s.MockClient{})
	if !reflect.DeepEqual(err, expErr) {
		t.Errorf("Expected: [%s], Got: [%s]", expErr, err)
	}
}

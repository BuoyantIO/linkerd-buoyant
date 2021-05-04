package k8s

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNewClient(t *testing.T) {
	fixtures := []*struct {
		testName string
		objects  []runtime.Object
	}{
		{
			"no objects",
			nil,
		},
		{
			"one pod",
			[]runtime.Object{
				&corev1.Pod{},
			},
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			c := fakeClient(tc.objects...)
			c.Sync(nil, time.Second)
		})
	}
}

func TestSync(t *testing.T) {
	openCh := make(chan struct{})
	bufferedCh := make(chan struct{}, 10)
	closedCh := make(chan struct{})
	close(closedCh)

	fixtures := []*struct {
		testName string
		stopCh   <-chan struct{}
		err      error
	}{
		{
			"nil channel",
			nil,
			nil,
		},
		{
			"open channel",
			openCh,
			nil,
		},
		{
			"buffered channel",
			bufferedCh,
			nil,
		},
		{
			"closed channel",
			closedCh,
			errSyncCache,
		},
	}

	for _, tc := range fixtures {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			c := fakeClient()
			err := c.Sync(tc.stopCh, time.Second)
			if err != tc.err {
				t.Errorf("Expected %s, got %s", tc.err, err)
			}
		})
	}
}

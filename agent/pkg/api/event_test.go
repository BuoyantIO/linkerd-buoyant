package api

import (
	"errors"
	"testing"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func TestAddEvent(t *testing.T) {
	t.Run("calls the api and gets a response", func(t *testing.T) {
		fixtures := []*struct {
			testName string
			event    *pb.Event
			err      error
		}{
			{
				"bad API response",
				&pb.Event{},
				errors.New("bad response"),
			},
			{
				"empty event",
				&pb.Event{},
				nil,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{err: tc.err}
				c := NewClient(m)

				err := c.AddEvent(tc.event)
				if tc.err != err {
					t.Errorf("Expected %s, got %s", tc.err, err)
				}

				if len(m.Events()) != 1 {
					t.Errorf("Expected 1 event, got %d", len(m.Events()))
				}
			})
		}
	})
}

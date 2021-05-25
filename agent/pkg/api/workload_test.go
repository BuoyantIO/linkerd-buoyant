package api

import (
	"testing"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func TestWorkloadStream(t *testing.T) {
	t.Run("streams a workload message", func(t *testing.T) {
		fixtures := []*struct {
			testName string
			msgs     []*pb.WorkloadMessage
			msgCount int
		}{
			{
				"no workload messages",
				nil,
				0,
			},
			{
				"one workload message",
				[]*pb.WorkloadMessage{{}},
				2,
			},
			{
				"two workload message",
				[]*pb.WorkloadMessage{{}, {}},
				3,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{}
				c := NewClient("", "", m)

				for _, msg := range tc.msgs {
					err := c.WorkloadStream(msg)
					if err != nil {
						t.Error(err)
					}
				}

				if len(m.Messages()) != tc.msgCount {
					t.Errorf("Expected %d message, got %d", tc.msgCount, len(m.Messages()))
				}
			})
		}
	})

	t.Run("sets auth info", func(t *testing.T) {
		fixtures := []*struct {
			testName string
			msgs     []*pb.WorkloadMessage
			expID    string
			expKey   string
		}{
			{
				"auth not set",
				nil,
				"",
				"",
			},
			{
				"auth set",
				[]*pb.WorkloadMessage{{}},
				fakeID,
				fakeKey,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{}
				c := NewClient(fakeID, fakeKey, m)

				for _, msg := range tc.msgs {
					err := c.WorkloadStream(msg)
					if err != nil {
						t.Error(err)
					}
				}

				if m.id != tc.expID {
					t.Errorf("Expected %s, got %s", tc.expID, m.id)
				}
				if m.key != tc.expKey {
					t.Errorf("Expected %s, got %s", tc.expKey, m.key)
				}
			})
		}
	})
}

package api

import (
	"errors"
	"testing"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func TestCrtInfo(t *testing.T) {
	t.Run("calls the api and gets a response", func(t *testing.T) {
		fixtures := []*struct {
			testName string
			info     *pb.CertificateInfo
			err      error
		}{
			{
				"bad API response",
				&pb.CertificateInfo{},
				errors.New("bad response"),
			},
			{
				"empty info",
				&pb.CertificateInfo{},
				nil,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{err: tc.err}
				c := NewClient(m)

				err := c.CrtInfo(tc.info)
				if tc.err != err {
					t.Errorf("Expected %s, got %s", tc.err, err)
				}

				if len(m.LinkerdMessages()) != 1 {
					t.Errorf("Expected 1 message, got %d", len(m.LinkerdMessages()))
				}
			})
		}
	})
}

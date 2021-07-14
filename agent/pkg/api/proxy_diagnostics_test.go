package api

import (
	"errors"
	"testing"
)

func TestProxyDiagnostic(t *testing.T) {
	t.Run("calls the api and gets a response", func(t *testing.T) {
		fixtures := []*struct {
			testName     string
			diagnosticId string
			logs         []byte
			metrics      []byte
			podManifest  []byte
			err          error
		}{
			{
				"bad API response",
				"diagnosticId",
				[]byte("logs"),
				[]byte("metrics"),
				[]byte("manifest"),
				errors.New("bad response"),
			},
			{
				"ok rsp",
				"diagnosticId",
				[]byte("logs"),
				[]byte("metrics"),
				[]byte("manifest"),
				nil,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{err: tc.err}
				c := NewClient("", "", m)

				err := c.ProxyDiagnostics(tc.diagnosticId, tc.logs, tc.metrics, tc.podManifest)
				if tc.err != err {
					t.Errorf("Expected %s, got %s", tc.err, err)
				}

				if len(m.ProxyDiagnosticMessages()) != 1 {
					t.Errorf("Expected 1 message, got %d", len(m.LinkerdMessages()))
				}
			})
		}
	})

	t.Run("sets auth info", func(t *testing.T) {
		m := &MockBcloudClient{}
		c := NewClient(fakeID, fakeKey, m)

		err := c.ProxyDiagnostics("id1", []byte{}, []byte{}, []byte{})
		if err != nil {
			t.Error(err)
		}

		if m.id != fakeID {
			t.Errorf("Expected %s, got %s", fakeID, m.id)
		}
		if m.key != fakeKey {
			t.Errorf("Expected %s, got %s", fakeKey, m.key)
		}
	})
}

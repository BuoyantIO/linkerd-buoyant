package api

import (
	"errors"
	"testing"
)

func TestProxyLogs(t *testing.T) {
	t.Run("calls the api and gets a response", func(t *testing.T) {
		fixtures := []*struct {
			testName     string
			podNamespace string
			podName      string
			logs         []byte
			err          error
		}{
			{
				"bad API response",
				"ns",
				"pod",
				[]byte("logs"),
				errors.New("bad response"),
			},
			{
				"ok rsp",
				"ns",
				"pod",
				[]byte("logs"),
				nil,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{err: tc.err}
				c := NewClient("", "", m)

				err := c.ProxyLogs(tc.podName, tc.podNamespace, tc.logs)
				if tc.err != err {
					t.Errorf("Expected %s, got %s", tc.err, err)
				}

				if len(m.ProxyLogsMessages()) != 1 {
					t.Errorf("Expected 1 message, got %d", len(m.ProxyDiagnosticMessages()))
				}
			})
		}
	})

	t.Run("sets auth info", func(t *testing.T) {
		m := &MockBcloudClient{}
		c := NewClient(fakeID, fakeKey, m)

		err := c.ProxyLogs("pod", "ns", nil)
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

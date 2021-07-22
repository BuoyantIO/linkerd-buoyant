package api

import (
	"errors"
	"testing"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func TestProxyDiagnostic(t *testing.T) {
	t.Run("calls the api and gets a response", func(t *testing.T) {
		fixtures := []*struct {
			testName       string
			diagnosticId   string
			logs           []byte
			metrics        [][]byte
			podManifest    *pb.Pod
			configMap      *pb.ConfigMap
			nodes          []*pb.Node
			k8sSvcManifest *pb.Service
			err            error
		}{
			{
				"bad API response",
				"diagnosticId",
				[]byte("logs"),
				[][]byte{[]byte("snapshot1"), []byte("snapshot2")},
				&pb.Pod{Pod: []byte("pod")},
				&pb.ConfigMap{ConfigMap: []byte("cm")},
				[]*pb.Node{{Node: []byte("node1")}, {Node: []byte("node2")}},
				&pb.Service{Service: []byte("svc")},
				errors.New("bad response"),
			},
			{
				"ok rsp",
				"diagnosticId",
				[]byte("logs"),
				[][]byte{[]byte("snapshot1"), []byte("snapshot2")},
				&pb.Pod{Pod: []byte("pod")},
				&pb.ConfigMap{ConfigMap: []byte("cm")},
				[]*pb.Node{{Node: []byte("node1")}, {Node: []byte("node2")}},
				&pb.Service{Service: []byte("svc")},
				nil,
			},
		}

		for _, tc := range fixtures {
			tc := tc
			t.Run(tc.testName, func(t *testing.T) {
				m := &MockBcloudClient{err: tc.err}
				c := NewClient("", "", m)

				err := c.ProxyDiagnostics(tc.diagnosticId, tc.logs, tc.metrics, tc.podManifest, tc.configMap, tc.nodes, tc.k8sSvcManifest)
				if tc.err != err {
					t.Errorf("Expected %s, got %s", tc.err, err)
				}

				if len(m.ProxyDiagnosticMessages()) != 1 {
					t.Errorf("Expected 1 message, got %d", len(m.ProxyDiagnosticMessages()))
				}
			})
		}
	})

	t.Run("sets auth info", func(t *testing.T) {
		m := &MockBcloudClient{}
		c := NewClient(fakeID, fakeKey, m)

		err := c.ProxyDiagnostics("id1", nil, nil, nil, nil, nil, nil)
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

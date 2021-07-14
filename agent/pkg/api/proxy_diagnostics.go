package api

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"google.golang.org/protobuf/encoding/prototext"
)

// AddEvent wraps the Buoyant Cloud API AddEvent gRPC unary endpoint.
func (c *Client) ProxyDiagnostics(diagnosticId string, logs []byte, metrics []byte, podManifest []byte) error {
	diagnostic := &pb.ProxyDiagnostic{
		Auth:         c.auth,
		DiagnosticId: diagnosticId,
		Logs:         logs,
		Metrics:      metrics,
		PodManifest:  podManifest,
	}
	c.log.Tracef("ProxyDiagnostics: %s", prototext.Format(diagnostic))

	_, err := c.client.ProxyDiagnostics(context.Background(), diagnostic)
	return err
}

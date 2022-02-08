package api

import (
	"context"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProxyLogs wraps the Buoyant Cloud API ProxyLogs gRPC unary endpoint.
func (c *Client) ProxyLogs(
	podName string,
	namespace string,
	logs []byte) error {
	now := time.Now()
	logMessage := &pb.ProxyLog{
		PodName:      podName,
		PodNamespace: namespace,
		Lines:        logs,
		Timestamp: &timestamppb.Timestamp{
			Seconds: now.Unix(),
			Nanos:   int32(now.Nanosecond()),
		},
	}
	c.log.Tracef("ProxyLogs: %s", prototext.Format(logMessage))

	_, err := c.client.ProxyLogs(context.Background(), logMessage)
	return err
}

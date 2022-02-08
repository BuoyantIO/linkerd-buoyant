package api

import (
	"context"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	"google.golang.org/protobuf/encoding/prototext"
)

// AddEvent wraps the Buoyant Cloud API AddEvent gRPC unary endpoint.
func (c *Client) AddEvent(event *pb.Event) error {
	bcloudEvent := &pb.Event{
		Event: event.Event,
		Owner: event.Owner,
	}
	c.log.Tracef("AddEvent: %s", prototext.Format(bcloudEvent))

	_, err := c.client.AddEvent(context.Background(), bcloudEvent)
	return err
}

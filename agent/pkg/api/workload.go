package api

import (
	"context"
	"io"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

// WorkloadStream wraps the Buoyant Cloud API WorkloadStream gRPC endpoint, and
// manages the client-side stream.
func (c *Client) WorkloadStream(msg *pb.WorkloadMessage) error {
	// loop and reset the stream if it has been closed
	for {
		err := c.sendMessage(msg)
		if err == io.EOF {
			c.log.Info("WorkloadStream closed")
			c.resetStream()
			continue
		} else if err != nil {
			c.log.Errorf("WorkloadStream failed to send: %s", err)
		}

		return err
	}
}

func (c *Client) sendMessage(msg *pb.WorkloadMessage) error {
	c.Lock()
	defer c.Unlock()
	if c.stream == nil {
		c.stream = c.newStream()
	}

	return c.stream.Send(msg)
}

func (c *Client) newStream() pb.Api_WorkloadStreamClient {
	var stream pb.Api_WorkloadStreamClient

	// loop until the request to initiate a stream succeeds
	for {
		var err error
		stream, err = c.client.WorkloadStream(context.Background())
		if err != nil {
			c.log.Errorf("failed to initiate stream: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		c.log.Info("WorkloadStream opened")

		err = stream.Send(&pb.WorkloadMessage{
			Message: &pb.WorkloadMessage_Auth{
				Auth: c.auth,
			},
		})
		if err != nil {
			c.log.Errorf("failed to send auth message: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}

	c.log.Info("WorkloadStream connected")
	return stream
}

func (c *Client) resetStream() {
	c.Lock()
	defer c.Unlock()
	if c.stream != nil {
		c.stream.CloseSend()
		c.stream = nil
	}
}

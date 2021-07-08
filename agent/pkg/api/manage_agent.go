package api

import (
	"context"
	"io"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
)

func (c *Client) RecvCommand() (*pb.AgentCommand, error) {
	for {
		if c.manageAgentClient == nil {
			c.manageAgentClient = c.newManageAgentClient()
		}

		msg, err := c.manageAgentClient.Recv()
		if err == io.EOF {
			c.log.Info("ManageAgentStream closed")
			c.manageAgentClient = nil
			continue
		} else if err != nil {
			c.log.Errorf("ManageAgentClient failed to receive: %s", err)
			return nil, err
		}
		return msg, nil
	}
}

func (c *Client) newManageAgentClient() pb.Api_ManageAgentClient {
	var client pb.Api_ManageAgentClient

	// loop until the request to initialize a client succeeds
	for {
		var err error
		client, err = c.client.ManageAgent(context.Background(), c.auth)
		if err != nil {
			c.log.Errorf("failed to initialize client: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		c.log.Info("ManageAgentClient initialized")

		break
	}

	c.log.Info("ManageAgentClient connected")
	return client
}

package api

import (
	"sync"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

// Client provides a simple Buoyant Cloud API client.
type Client struct {
	client            pb.ApiClient
	workloadStream    *workloadStream
	manageAgentStream *manageAgentStream

	log *log.Entry

	// protects stream
	sync.Mutex
}

// NewClient instantiates a new Buoyant Cloud API client.
func NewClient(client pb.ApiClient) *Client {
	return &Client{
		workloadStream:    newWorkloadStream(client),
		manageAgentStream: newManageAgentStream(client),
		client:            client,
		log:               log.WithField("api", "some-id"),
	}
}

func (c *Client) Start() {
	c.manageAgentStream.startStream()
}

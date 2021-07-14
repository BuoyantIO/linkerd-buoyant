package api

import (
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

// manageAgentStream wraps the Buoyant Cloud API ManageAgent gRPC endpoint, and
// manages the stream.
type manageAgentStream struct {
	auth   *pb.Auth
	client pb.ApiClient
	stream pb.Api_ManageAgentClient
	log    *log.Entry

	commands chan *pb.AgentCommand

	// protects stream
	sync.Mutex
}

func newManageAgentStream(auth *pb.Auth, client pb.ApiClient) *manageAgentStream {
	return &manageAgentStream{
		auth:     auth,
		client:   client,
		log:      log.WithField("stream", "ManageAgentStream"),
		commands: make(chan *pb.AgentCommand, 10),
	}
}

func (s *manageAgentStream) startStream() {
	for {
		command, err := s.recv_locked()
		if err != nil {
			if err == io.EOF {
				s.log.Info("server closed stream, reseting")
			}
			s.resetStream()
			continue
		}
		s.commands <- command
	}
}

func (s *manageAgentStream) recv_locked() (*pb.AgentCommand, error) {
	s.Lock()
	defer s.Unlock()
	if s.stream == nil {
		s.stream = s.newStream()
	}

	return s.stream.Recv()
}

func (s *manageAgentStream) newStream() pb.Api_ManageAgentClient {
	var stream pb.Api_ManageAgentClient

	// loop until the request to initiate a stream succeeds
	for {
		var err error
		stream, err = s.client.ManageAgent(context.Background(), s.auth)
		if err != nil {
			s.log.Errorf("failed to initiate stream: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	s.log.Info("ManageAgentStream connected")
	return stream
}

func (s *manageAgentStream) resetStream() {
	s.Lock()
	defer s.Unlock()
	if s.stream != nil {
		s.stream.CloseSend()
		s.stream = nil
	}
}

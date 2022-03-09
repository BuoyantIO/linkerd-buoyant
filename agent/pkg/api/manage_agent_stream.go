package api

import (
	"context"
	"sync"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

// manageAgentStream wraps the Buoyant Cloud API ManageAgent gRPC endpoint, and
// manages the stream.
type manageAgentStream struct {
	client pb.ApiClient
	stream pb.Api_ManageAgentClient
	log    *log.Entry

	commands chan *pb.AgentCommand

	// protects stream
	sync.Mutex
}

func newManageAgentStream(client pb.ApiClient) *manageAgentStream {
	return &manageAgentStream{
		client:   client,
		log:      log.WithField("stream", "ManageAgentStream"),
		commands: make(chan *pb.AgentCommand),
	}
}

func (s *manageAgentStream) startStream() {
	for {
		command, err := s.recvCommand()
		if err != nil {
			// TODO: set this back to `Infof`:
			// https://github.com/BuoyantIO/linkerd-buoyant/issues/21
			s.log.Debugf("stream closed, reseting: %s", err)
			s.resetStream()
			continue
		}
		s.commands <- command
	}
}

func (s *manageAgentStream) recvCommand() (*pb.AgentCommand, error) {
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
		stream, err = s.client.ManageAgent(context.Background(), &pb.Auth{})
		if err != nil {
			s.log.Errorf("failed to initiate stream: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}

	// TODO: set this back to `Info`:
	// https://github.com/BuoyantIO/linkerd-buoyant/issues/21
	s.log.Debug("ManageAgentStream connected")
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

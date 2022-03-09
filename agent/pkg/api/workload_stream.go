package api

import (
	"context"
	"io"
	"sync"
	"time"

	pb "github.com/buoyantio/linkerd-buoyant/gen/bcloud"
	log "github.com/sirupsen/logrus"
)

// WorkloadStream wraps the Buoyant Cloud API WorkloadStream gRPC endpoint, and
// manages the client-side stream.
type workloadStream struct {
	client pb.ApiClient
	stream pb.Api_WorkloadStreamClient
	log    *log.Entry

	// protects stream
	sync.Mutex
}

func newWorkloadStream(client pb.ApiClient) *workloadStream {
	return &workloadStream{
		client: client,
		log:    log.WithField("stream", "WorkloadStream"),
	}
}

func (s *workloadStream) send(msg *pb.WorkloadMessage) error {
	// loop and reset the stream if it has been closed
	for {
		err := s.sendMessage(msg)
		if err == io.EOF {
			s.log.Info("WorkloadStream closed")
			s.resetStream()
			continue
		} else if err != nil {
			s.log.Errorf("WorkloadStream failed to send: %s", err)
		}

		return err
	}
}

func (s *workloadStream) sendMessage(msg *pb.WorkloadMessage) error {
	s.Lock()
	defer s.Unlock()
	if s.stream == nil {
		s.stream = s.newStream()
	}

	return s.stream.Send(msg)
}

func (s *workloadStream) newStream() pb.Api_WorkloadStreamClient {
	var stream pb.Api_WorkloadStreamClient

	// loop until the request to initiate a stream succeeds
	for {
		var err error
		stream, err = s.client.WorkloadStream(context.Background())
		if err != nil {
			s.log.Errorf("failed to initiate stream: %s", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		s.log.Info("WorkloadStream opened")

		break
	}

	s.log.Info("WorkloadStream connected")
	return stream
}

func (s *workloadStream) resetStream() {
	s.Lock()
	defer s.Unlock()
	if s.stream != nil {
		s.stream.CloseSend()
		s.stream = nil
	}
}

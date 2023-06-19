package handler

import (
	"context"

	log "github.com/micro/micro/v3/service/logger"

	match_process "match_process/proto"
)

type Match_process struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_process) Call(ctx context.Context, req *match_process.Request, rsp *match_process.Response) error {
	log.Info("Received Match_process.Call request")
	rsp.Msg = "Hello " + req.Name
	return nil
}

// Stream is a server side stream handler called via client.Stream or the generated client code
func (e *Match_process) Stream(ctx context.Context, req *match_process.StreamingRequest, stream match_process.Match_process_StreamStream) error {
	log.Infof("Received Match_process.Stream request with count: %d", req.Count)

	for i := 0; i < int(req.Count); i++ {
		log.Infof("Responding: %d", i)
		if err := stream.Send(&match_process.StreamingResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
	}

	return nil
}

// PingPong is a bidirectional stream handler called via client.Stream or the generated client code
func (e *Match_process) PingPong(ctx context.Context, stream match_process.Match_process_PingPongStream) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		log.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&match_process.Pong{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}

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

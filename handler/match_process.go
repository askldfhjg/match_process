package handler

import (
	"context"
	match_process "match_process/proto"
)

type Match_process struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_process) MatchTask(ctx context.Context, req *match_process.MatchTaskReq, rsp *match_process.MatchTaskRsp) error {
	return nil
}

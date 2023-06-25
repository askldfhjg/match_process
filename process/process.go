package process

import match_evaluator "github.com/askldfhjg/match_apis/match_evaluator/proto"

var DefaultManager Manager

type Manager interface {
	Start() error
	Stop() error
	AddEvalOpt(req *match_evaluator.ToEvalReq, key string)
}

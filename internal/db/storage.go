package db

import (
	"context"

	match_process "match_process/proto"

	match_frontend "github.com/askldfhjg/match_apis/match_frontend/proto"
)

var Default Service

type Service interface {
	Init(ctx context.Context, opts ...Option) error
	Close(ctx context.Context) error
	String() string
	GetTokenList(ctx context.Context, info *match_process.MatchTaskReq) ([]string, error)
	GetTokenDetail(ctx context.Context, ids []string) ([]*match_frontend.MatchInfo, error)
	SetEvalUrl(ctx context.Context, hashkey string, url string) (string, error)
	RemoveMissTokens(ctx context.Context, playerIds []string, gameId string, subType int64) (int, error)
}

type MatchInfo struct {
	Id       string
	PlayerId string
	Score    int64
	GameId   string
	SubType  int64
}

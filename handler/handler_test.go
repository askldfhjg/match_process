package handler

import (
	"context"
	"testing"

	"match_process/internal/db"
	"match_process/internal/db/redis"
	match_process "match_process/proto"
)

//	func BenchmarkMatchTask67(b *testing.B) {
//		svr, _ := redis.New(
//			db.WithAddress("127.0.0.1:6379"),
//			db.WithPoolMaxActive(5),
//			db.WithPoolMaxIdle(100),
//			db.WithPoolIdleTimeout(300))
//		db.Default = svr
//		ids := []string{
//			"aa",
//			"bb",
//			"cc",
//			"dd",
//			"ee",
//			"ff",
//		}
//		for n := 0; n < b.N; n++ {
//			db.Default.GetTokenDetail(context.Background(), ids)
//		}
//	}
func BenchmarkMatchTask(b *testing.B) {
	// create handler instance
	svr, _ := redis.New(
		db.WithAddress("127.0.0.1:6379"),
		db.WithPoolMaxActive(5),
		db.WithPoolMaxIdle(100),
		db.WithPoolIdleTimeout(300))
	db.Default = svr
	handler := &Match_process{}
	req := &match_process.MatchTaskReq{
		TaskId:             "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375",
		SubTaskId:          "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375-42",
		GameId:             "aaaa",
		SubType:            1,
		StartPos:           10000,
		EndPos:             12200,
		EvalGroupId:        "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375",
		EvalGroupTaskCount: 44,
		EvalGroupSubId:     18,
		EvalhaskKey:        "HnypVaXhQyMKRFO",
		NeedCount:          3,
		Version:            1687761589372360938,
	}
	for n := 0; n < b.N; n++ {
		// Context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Call the handler
		rsp := &match_process.MatchTaskRsp{}
		handler.MatchTask(ctx, req, rsp)
	}

}

func TestMatchTask(t *testing.T) {
	// create handler instance
	svr, _ := redis.New(
		db.WithAddress("127.0.0.1:6379"),
		db.WithPoolMaxActive(5),
		db.WithPoolMaxIdle(100),
		db.WithPoolIdleTimeout(300))
	db.Default = svr
	handler := &Match_process{}
	req := &match_process.MatchTaskReq{
		TaskId:             "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375",
		SubTaskId:          "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375-42",
		GameId:             "aaaa",
		SubType:            1,
		StartPos:           10000,
		EndPos:             12200,
		EvalGroupId:        "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375",
		EvalGroupTaskCount: 44,
		EvalGroupSubId:     18,
		EvalhaskKey:        "HnypVaXhQyMKRFO",
		NeedCount:          3,
		Version:            1687761589372360938,
	}

	// Context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Call the handler
	rsp := &match_process.MatchTaskRsp{}
	handler.MatchTask(ctx, req, rsp)
}

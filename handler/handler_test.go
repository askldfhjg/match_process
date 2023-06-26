package handler

import (
	"context"
	"math/rand"
	"testing"

	"match_process/internal/db"
	"match_process/internal/db/redis"
	match_process "match_process/proto"

	"github.com/stretchr/testify/assert"
)

func TestMatchTask(t *testing.T) {
	// create handler instance
	svr, err := redis.New(
		db.WithAddress("127.0.0.1:6379"),
		db.WithPoolMaxActive(5),
		db.WithPoolMaxIdle(100),
		db.WithPoolIdleTimeout(300))
	assert.NoError(t, err)
	db.Default = svr
	handler := &Match_process{}

	// Prepare request
	req := &match_process.MatchTaskReq{
		TaskId:             "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375",
		SubTaskId:          "d44a388b-0874-4b39-88c1-989971cbfdd0-1687761589375-42",
		GameId:             "aaaa",
		SubType:            1,
		StartPos:           33901,
		EndPos:             36100,
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
	err = handler.MatchTask(ctx, req, rsp)
	assert.NoError(t, err)
	assert.NotNil(t, rsp)

	// Validate response fields
	//assert.NotEmpty(t, rsp.EvalhaskKey)
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

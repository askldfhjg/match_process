package redis

import (
	"bytes"
	"context"
	"fmt"

	match_process "match_process/proto"
	"match_process/utils"

	"github.com/gomodule/redigo/redis"
)

const (
	allTickets = "allTickets:%d:%s:%d"
	ticketKey  = "ticket:"
	hashKey    = "hashkey:"
)

func (m *redisBackend) GetTokenList(ctx context.Context, info *match_process.MatchTaskReq) ([]interface{}, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)

	zsetKey := fmt.Sprintf(allTickets, info.OldVersion, info.GameId, info.SubType)
	reply, err := redisConn.Do("ZRANGEBYSCORE", zsetKey, info.StartPos, info.EndPos, "WITHSCORES")
	if err == redis.ErrNil {
		return nil, nil
	}
	return reply.([]interface{}), nil
}

func (m *redisBackend) GetTokenDetail(ctx context.Context, ids []string) ([]interface{}, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)

	queryParams := make([]interface{}, len(ids))
	tmps := make([]*bytes.Buffer, len(ids))
	for i, id := range ids {
		by := utils.GetBytes()
		by.WriteString(ticketKey)
		by.WriteString(id)
		bbb := utils.Bytes2string(by.Bytes())
		queryParams[i] = bbb
		tmps[i] = by
	}
	queryParams, err = redis.Values(redisConn.Do("MGET", queryParams...))
	if err != nil {
		return nil, err
	}
	for _, it := range tmps {
		utils.PutBytes(it)
	}
	return queryParams, nil
}

func (m *redisBackend) SetEvalUrl(ctx context.Context, hashkey string, url string) (string, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return "", err
	}
	defer handleConnectionClose(&redisConn)

	script := `local oldV = redis.call('GET', KEYS[1])
	if(not oldV) then
		redis.call('SET', KEYS[1], ARGV[1], 'EX', 10)
		return ARGV[1]
	else
		return oldV
	end
	`
	args := []interface{}{url}
	keys := []interface{}{hashKey + hashkey}
	params := []interface{}{script, len(keys)}
	params = append(params, keys...)
	params = append(params, args...)
	return redis.String(redisConn.Do("EVAL", params...))
}

// func (m *redisBackend) RemoveMissTokens(ctx context.Context, playerIds []string, gameId string, subType int64) (int, error) {
// 	redisConn, err := m.redisPool.GetContext(ctx)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer handleConnectionClose(&redisConn)
// 	zsetKey := fmt.Sprintf(allTickets, gameId, subType)

// 	inter2 := make([]interface{}, len(playerIds)+1)
// 	inter2[0] = zsetKey
// 	for pos, ply := range playerIds {
// 		inter2[pos+1] = ply
// 	}
// 	return redis.Int(redisConn.Do("ZREM", inter2...))
// }

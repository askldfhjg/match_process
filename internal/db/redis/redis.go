package redis

import (
	"context"
	"fmt"

	match_process "match_process/proto"

	match_frontend "github.com/askldfhjg/match_apis/match_frontend/proto"

	"github.com/gomodule/redigo/redis"
	"google.golang.org/protobuf/proto"
)

func (m *redisBackend) GetTokenList(ctx context.Context, info *match_process.MatchTaskReq) ([]string, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)

	zsetKey := fmt.Sprintf(allTickets, info.GameId, info.SubType)
	reply, err := redisConn.Do("ZRANGE", zsetKey, info.StartPos-1, info.EndPos-1)
	if err == redis.ErrNil {
		return nil, nil
	}
	v, err := redis.Values(reply, err)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0)
	if err = redis.ScanSlice(v, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (m *redisBackend) GetTokenDetail(ctx context.Context, ids []string) ([]*match_frontend.MatchInfo, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer handleConnectionClose(&redisConn)

	queryParams := make([]interface{}, len(ids))
	for i, id := range ids {
		queryParams[i] = fmt.Sprintf(ticketKey, id)
	}

	ticketBytes, err := redis.ByteSlices(redisConn.Do("MGET", queryParams...))

	if err != nil {
		return nil, err
	}

	r := make([]*match_frontend.MatchInfo, 0, len(queryParams))
	for _, b := range ticketBytes {
		if b != nil {
			t := &match_frontend.MatchInfo{}
			err = proto.Unmarshal(b, t)
			if err != nil {
				continue
			}
			r = append(r, t)
		}
	}
	return r, nil
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
	keys := []interface{}{fmt.Sprintf(hashKey, hashkey)}
	params := []interface{}{script, len(keys)}
	params = append(params, keys...)
	params = append(params, args...)
	return redis.String(redisConn.Do("EVAL", params...))
}

func (m *redisBackend) RemoveMissTokens(ctx context.Context, playerIds []string, gameId string, subType int64) (int, error) {
	redisConn, err := m.redisPool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer handleConnectionClose(&redisConn)
	zsetKey := fmt.Sprintf(allTickets, gameId, subType)
	inter2 := make([]interface{}, 0, len(playerIds))
	inter2 = append(inter2, zsetKey)
	for _, ply := range playerIds {
		inter2 = append(inter2, ply)
	}
	return redis.Int(redisConn.Do("ZREM", inter2...))
}

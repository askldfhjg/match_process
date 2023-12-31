package handler

import (
	"context"
	"fmt"
	"match_process/internal/db"
	"match_process/process"
	match_process "match_process/proto"
	"match_process/utils"
	"math"
	"strconv"
	"time"

	match_evaluator "github.com/askldfhjg/match_apis/match_evaluator/proto"
	"github.com/micro/micro/v3/service/logger"
)

const detailOnceGetCount = 300
const scoreMaxOffset = 100

type Match_process struct{}

type matchList struct {
	list     []string
	scoreMap map[string]float64
	index    int
	result   map[string]bool
	gameId   string
	subType  int64
	SubId    int64
	nowMatch string
}

func newMatchList(req *match_process.MatchTaskReq) (*matchList, error) {
	li, err := db.Default.GetTokenList(context.Background(), req)
	if err != nil {
		return nil, err
	}

	allCount := len(li)
	playerList := make([]string, 0, allCount/2+1)
	scoreMap := make(map[string]float64, allCount/2+1)
	index := 0
	for {
		if index+1 >= allCount {
			break
		}
		s1, ok := li[index].([]byte)
		if !ok {
			index += 2
			continue
		}
		s2, ok := li[index+1].([]byte)
		if !ok {
			index += 2
			continue
		}
		playerId := utils.Bytes2string(s1)
		scoreStr := utils.Bytes2string(s2)
		score, err := strconv.ParseInt(scoreStr, 10, 64)
		if err == nil {
			playerList = append(playerList, playerId)
			scoreMap[playerId] = float64(score)
		}
		index += 2
	}
	//logger.Infof("ffffff %+v", len(scoreMap))
	// for _, info := range li {
	// 	playerId := info.Member.(string)
	// 	score := info.Score
	// 	playerList = append(playerList, playerId)
	// 	scoreMap[playerId] = score
	// }

	return &matchList{
		list:     playerList,
		scoreMap: scoreMap,
		result:   make(map[string]bool, allCount/2+1),
		gameId:   req.GameId,
		subType:  req.SubType,
		SubId:    req.EvalGroupSubId,
		nowMatch: fmt.Sprintf("%s:%d", req.GameId, req.SubType),
	}, nil
}

func (m *matchList) GetDetail(i int) (float64, bool) {
	if i >= len(m.list) {
		return 0, false
	}
	if i >= m.index {
		st := m.index
		if st < 0 {
			st = 0
		}
		ed := st + detailOnceGetCount
		if ed > len(m.list) {
			ed = len(m.list)
		}
		rr, err := db.Default.GetTokenDetail(context.Background(), m.list[st:ed])
		//logger.Infof("GetTokenDetail %d %d", st, ed)
		//deleteIds := make([]string, 0, 64)
		//var deleteIds []string
		m.index = ed
		if err == nil {
			for pos, idd := range m.list[st:ed] {
				infer := rr[pos]
				if infer == nil {
					//deleteIds = append(deleteIds, idd)
					continue
				}
				str, ok := rr[pos].([]byte)
				if !ok {
					//deleteIds = append(deleteIds, idd)
					continue
				}
				matchInfo := utils.Bytes2string(str)
				if matchInfo == m.nowMatch {
					m.result[idd] = true
				} else {
					//deleteIds = append(deleteIds, idd)
				}
			}
			// if len(deleteIds) > 0 {
			// 	logger.Infof("GetDetail %s %d %d remove miss count %d range %d %d", m.gameId, m.subType, m.SubId, len(deleteIds), st, ed)
			// 	_, err := db.Default.RemoveMissTokens(context.Background(), deleteIds, m.gameId, m.subType)
			// 	if err != nil {
			// 		logger.Errorf("GetDetail %s %d $d remove miss have error %s", m.gameId, m.subType, m.SubId, err.Error())
			// 	}
			// 	// else {
			// 	// 	if deleteCount != len(deleteIds) {
			// 	// 		logger.Errorf("GetDetail %s %d %d remove miss delete not match %d %d range %d %d", m.gameId, m.subType, m.SubId, deleteCount, len(deleteIds), st, ed)
			// 	// 	}
			// 	// }
			// }
		} else {
			logger.Errorf("GetDetail redis error %s", err.Error())
		}
	}
	item := m.list[i]
	return m.scoreMap[item], m.result[item]
}

func (m matchList) Count() int {
	return len(m.list)
}

func (m matchList) GetPlayerIds(poss []int, maxCount int) []string {
	rr := make([]string, maxCount)
	for i := 0; i < maxCount; i++ {
		if i < len(poss) {
			rr[i] = m.list[poss[i]]
		} else {
			rr[i] = "robot"
		}
	}
	return rr
}

func (m matchList) GetIds(poss []int) map[string]int32 {
	rr := make(map[string]int32)
	for _, pos := range poss {
		idd := m.list[pos]
		score := m.scoreMap[idd]
		if score > 0 {
			rr[idd] = int32(score)
		}
	}
	return rr
}

func (m matchList) CalcScore(poss []int) int64 {
	currentGroupMin := math.MaxFloat64
	currentGroupMax := -1.0
	for _, pos := range poss {
		score, ok := m.GetDetail(pos)
		if !ok {
			continue
		}
		currentGroupMin = math.Min(currentGroupMin, score)
		currentGroupMax = math.Max(currentGroupMax, score)
	}
	ret := currentGroupMax - currentGroupMin
	if ret < 0 {
		ret = 0
	}
	return int64(ret)
}

func groupWithinOffsetAndMaxCount(playerIds []int, mList *matchList, maxOffset, maxCount int, robotCount int, gameId string) ([]*match_evaluator.MatchDetail, []int) {
	groups := make([]*match_evaluator.MatchDetail, 0, 64)
	remind := make([]int, 0, 64)
	currentGroup := make([]int, 0, maxCount)
	currentGroupMin := math.MaxFloat64
	currentGroupMax := -1.0

	groupsWithinMaxOffset := func(min float64, max float64, newV float64, maxOffset int) bool {
		min = math.Min(min, newV)
		max = math.Max(max, newV)
		return max-min < float64(maxOffset)
	}

	processResult := func(tmp []int, mList *matchList, maxCount int) *match_evaluator.MatchDetail {
		ret := &match_evaluator.MatchDetail{
			Ids: mList.GetPlayerIds(tmp, maxCount),
		}
		return ret
	}

	for _, pos := range playerIds {
		score, ok := mList.GetDetail(pos)
		if !ok {
			continue
		}
		if len(currentGroup) == 0 {
			currentGroup = append(currentGroup, pos)
			currentGroupMin = math.Min(currentGroupMin, score)
			currentGroupMax = math.Max(currentGroupMax, score)
		} else if len(currentGroup) < maxCount && groupsWithinMaxOffset(currentGroupMin, currentGroupMax, score, maxOffset) {
			currentGroup = append(currentGroup, pos)
			currentGroupMin = math.Min(currentGroupMin, score)
			currentGroupMax = math.Max(currentGroupMax, score)
		} else {
			cc := len(currentGroup)
			if cc >= maxCount {
				groups = append(groups, processResult(currentGroup, mList, maxCount))
			} else if cc >= robotCount {
				groups = append(groups, processResult(currentGroup, mList, maxCount))
			} else {
				remind = append(remind, currentGroup...)
			}
			currentGroup = currentGroup[:0]
			currentGroup = append(currentGroup, pos)
			currentGroupMin = score
			currentGroupMax = score
		}
	}

	cc := len(currentGroup)
	if cc >= maxCount {
		groups = append(groups, processResult(currentGroup, mList, maxCount))
	} else if cc >= robotCount {
		groups = append(groups, processResult(currentGroup, mList, maxCount))
	} else {
		remind = append(remind, currentGroup...)
	}
	return groups, remind
}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_process) MatchTask(ctx context.Context, req *match_process.MatchTaskReq, rsp *match_process.MatchTaskRsp) error {
	//logger.Info("Received MatchProcess.Call request")
	st := time.Now().UnixNano() / 1e6
	mList, err := newMatchList(req)
	if err != nil {
		rsp.Code = -1
		rsp.Err = err.Error()
		return err
	}
	cc := mList.Count()
	tmpList := make([]int, cc)
	for i := 0; i < cc; i++ {
		tmpList[i] = i
	}
	// evalReadyReq := &match_evaluator.ToEvalReadyReq{
	// 	GameId:             req.GameId,
	// 	SubType:            req.SubType,
	// 	Version:            req.Version,
	// 	EvalGroupId:        req.EvalGroupId,
	// 	EvalGroupTaskCount: req.EvalGroupTaskCount,
	// 	EvalGroupSubId:     req.EvalGroupSubId,
	// 	OldVersion:         req.OldVersion,
	// }
	// process.DefaultManager.AddEvalReadyOpt(evalReadyReq, req.EvalhaskKey)
	go func() {
		var remindold []int
		var ret1 []*match_evaluator.MatchDetail
		ret, remind := groupWithinOffsetAndMaxCount(tmpList, mList, scoreMaxOffset, int(req.NeedCount), int(req.NeedCount), req.GameId)
		if len(remind) > 0 {
			ret1, remindold = groupWithinOffsetAndMaxCount(remind, mList, scoreMaxOffset*3, int(req.NeedCount), int(float64(req.NeedCount)*0.8), req.GameId)
			ret = append(ret, ret1...)
		}
		now := time.Now().UnixNano() / 1e6
		logger.Infof("process %s %d ok count %d move %d timer %d %d", req.EvalGroupId, req.EvalGroupSubId, len(ret), len(remind), now, now-st)
		evalReq := &match_evaluator.ToEvalReq{
			Details:            ret,
			TaskId:             req.TaskId,
			SubTaskId:          req.SubTaskId,
			GameId:             req.GameId,
			SubType:            req.SubType,
			Version:            req.Version,
			EvalGroupId:        req.EvalGroupId,
			EvalGroupTaskCount: req.EvalGroupTaskCount,
			EvalGroupSubId:     req.EvalGroupSubId,
			StartTime:          req.StartTime,
			RunTime:            now - st,
			OldVersion:         req.OldVersion,
			MoveList:           mList.GetIds(remindold),
		}
		process.DefaultManager.AddEvalOpt(evalReq, req.EvalhaskKey)
	}()

	return nil
}

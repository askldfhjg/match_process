package handler

import (
	"context"
	"match_process/internal/db"
	"match_process/process"
	match_process "match_process/proto"
	"math"
	"time"

	match_evaluator "github.com/askldfhjg/match_apis/match_evaluator/proto"
	match_frontend "github.com/askldfhjg/match_apis/match_frontend/proto"
	"github.com/micro/micro/v3/service/logger"
)

const detailOnceGetCount = 100
const scoreMaxOffset = 100

type Match_process struct{}

type matchList struct {
	list    []string
	index   int
	result  map[string]*match_frontend.MatchInfo
	gameId  string
	subType int64
}

func (m *matchList) GetDetail(i int) *match_frontend.MatchInfo {
	if i >= len(m.list) {
		return nil
	}
	if i >= m.index-10 {
		st := m.index - 10
		if st < 0 {
			st = 0
		}
		ed := st + detailOnceGetCount
		if ed > len(m.list) {
			ed = len(m.list)
		}
		rr, err := db.Default.GetTokenDetail(context.Background(), m.list[st:ed])
		//logger.Infof("GetTokenDetail %v", err)
		if err == nil {
			m.index = ed
			for _, info := range rr {
				if info.GameId == m.gameId && info.SubType == m.subType {
					m.result[info.PlayerId] = info
				}
			}
		}
	}
	return m.result[m.list[i]]
}

func (m matchList) GetPlayerIds(poss []int) []string {
	rr := make([]string, len(poss))
	for idx, pos := range poss {
		rr[idx] = m.list[pos]
	}
	return rr
}

func (m matchList) CalcScore(poss []int) int64 {
	currentGroupMin := math.MaxFloat64
	currentGroupMax := -1.0
	for _, pos := range poss {
		detail := m.GetDetail(pos)
		if detail == nil {
			continue
		}
		score := float64(detail.Score)
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

	processResult := func(tmp []int, mList *matchList, gameId string) *match_evaluator.MatchDetail {
		ret := &match_evaluator.MatchDetail{
			Ids:    mList.GetPlayerIds(tmp),
			GameId: gameId,
			Score:  mList.CalcScore(tmp),
		}
		return ret
	}

	for _, pos := range playerIds {
		detail := mList.GetDetail(pos)
		if detail == nil {
			continue
		}
		score := float64(detail.Score)
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
				groups = append(groups, processResult(currentGroup, mList, gameId))
			} else if cc >= robotCount {
				groups = append(groups, processResult(currentGroup, mList, gameId))
			} else {
				remind = append(remind, currentGroup...)
			}
			currentGroup = make([]int, 1, maxCount)
			currentGroup[0] = pos
			currentGroupMin = score
			currentGroupMax = score
		}
	}

	cc := len(currentGroup)
	if cc >= maxCount {
		groups = append(groups, processResult(currentGroup, mList, gameId))
	} else if cc >= robotCount {
		groups = append(groups, processResult(currentGroup, mList, gameId))
	} else {
		remind = append(remind, currentGroup...)
	}
	return groups, remind
}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_process) MatchTask(ctx context.Context, req *match_process.MatchTaskReq, rsp *match_process.MatchTaskRsp) error {
	logger.Info("Received MatchProcess.Call request")
	//rsp.Msg = "Hello " + req.Name
	li, err := db.Default.GetTokenList(context.Background(), req)
	if err != nil {
		rsp.Code = -1
		rsp.Err = err.Error()
		return err
	}
	mList := &matchList{
		list:    li,
		result:  make(map[string]*match_frontend.MatchInfo, 32),
		gameId:  req.GameId,
		subType: req.SubType,
	}
	tmpList := make([]int, len(li))
	for i := 0; i < len(li); i++ {
		tmpList[i] = i
	}
	go func() {
		ret, remind := groupWithinOffsetAndMaxCount(tmpList, mList, scoreMaxOffset, int(req.NeedCount), int(req.NeedCount), req.GameId)
		if len(remind) > 0 {
			ret1, _ := groupWithinOffsetAndMaxCount(remind, mList, scoreMaxOffset*3, int(req.NeedCount), int(float64(req.NeedCount)*0.8), req.GameId)
			ret = append(ret, ret1...)
		}
		logger.Infof("process %v %v ok count %v timer %v", req.EvalGroupId, req.EvalGroupSubId, len(ret), time.Now().UnixNano()/1e6)
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
		}
		process.DefaultManager.AddEvalOpt(evalReq, req.EvalhaskKey)
	}()

	return nil
}

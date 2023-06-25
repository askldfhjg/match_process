package handler

import (
	"context"
	"match_process/internal/db"
	match_process "match_process/proto"
	"math"

	match_evaluator "github.com/askldfhjg/match_apis/match_evaluator/proto"
	match_frontend "github.com/askldfhjg/match_apis/match_frontend/proto"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
)

const detailOnceGetCount = 100
const scoreMaxOffset = 100

type Match_process struct{}

type matchList struct {
	ctx    context.Context
	list   []string
	index  int
	result map[string]*match_frontend.MatchInfo
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
		rr, err := db.Default.GetTokenDetail(m.ctx, m.list[st:ed])
		if err != nil {
			m.index = ed
			for _, info := range rr {
				m.result[info.PlayerId] = info
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
	return 0
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
			currentGroup = []int{pos}
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
	li, err := db.Default.GetTokenList(ctx, req)
	if err != nil {
		return err
	}
	mList := &matchList{
		ctx:    ctx,
		list:   li,
		result: make(map[string]*match_frontend.MatchInfo, 32),
	}
	tmpList := make([]int, len(li))
	for i := 0; i < len(li); i++ {
		tmpList[i] = i
	}
	ret, remind := groupWithinOffsetAndMaxCount(tmpList, mList, scoreMaxOffset, int(req.NeedCount), int(req.NeedCount), req.GameId)
	if len(remind) > 0 {
		ret1, _ := groupWithinOffsetAndMaxCount(remind, mList, scoreMaxOffset*3, int(req.NeedCount), int(float64(req.NeedCount)*0.8), req.GameId)
		ret = append(ret, ret1...)
	}
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
	evalSrv := match_evaluator.NewMatchEvaluatorService("match_evaluator", client.DefaultClient)
	evalRsp, err := evalSrv.ToEval(context.Background(), evalReq)
	if err != nil {
		logger.Infof("ToEval error %+v", err)
	} else {
		logger.Infof("ToEval result %+v", evalRsp)
	}
	return nil
}

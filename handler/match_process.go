package handler

import (
	"container/list"
	"context"
	"match_process/internal/db"
	match_process "match_process/proto"

	match_frontend "github.com/askldfhjg/match_apis/match_frontend/proto"
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

func (m matchList) GetDetail(i int) *match_frontend.MatchInfo {
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
	stPos := 0
	edPos := 0
	isStart := false
	condition := list.New()
	for {
		if !isStart {
			detail := mList.GetDetail(stPos)
			if detail == nil {
				stPos++
				continue
			}
			condition.PushBack(stPos)
			isStart = true
			edPos = stPos + 1
		}
		detail := mList.GetDetail(edPos)
		if detail == nil {
			edPos++
			continue
		}
		if detail.Score-mList.GetDetail(stPos).Score < scoreMaxOffset {
			laPos := condition.Back().Value.(int)
			if laPos != edPos {
				condition.PushBack(edPos)
			}
			if condition.Len() >= int(req.NeedCount) {
				stPos = edPos + 1
				edPos = stPos
				isStart = false
				condition = list.New()
			} else {
				edPos++
			}
		} else {
			if condition.Len() > 1 {
				condition.Remove(condition.Front())
				stPos = condition.Front().Value.(int)
			} else {
				stPos = edPos
				edPos = stPos
				isStart = false
				condition = list.New()
			}
		}
	}
}

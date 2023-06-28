package manager

import (
	"context"
	"match_process/internal/consistenthash"
	"match_process/process"
	"os"

	match_evaluator "github.com/askldfhjg/match_apis/match_evaluator/proto"

	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
	"github.com/micro/micro/v3/service/registry"
)

func NewManager(opts ...process.ProcessOption) process.Manager {
	m := &defaultMgr{
		exited:      make(chan struct{}, 1),
		watcher:     nil,
		ring:        consistenthash.New(10, nil),
		nodeChannel: make(chan *nodeChangeOpt, 2000),
		evalChannel: make(chan *evalOpt, 2000),
	}
	for _, o := range opts {
		o(&m.opts)
	}
	return m
}

const (
	NodeOptUpdate int = 0
	NodeOptDelete     = 1
)

type evalOpt struct {
	Req *match_evaluator.ToEvalReq
	Key string
}
type nodeChangeOpt struct {
	Opt     int
	Address string
}

type defaultMgr struct {
	opts        process.ProcessOptions
	exited      chan struct{}
	watcher     registry.Watcher
	ring        *consistenthash.HashRing
	nodeChannel chan *nodeChangeOpt
	evalChannel chan *evalOpt
}

func (m *defaultMgr) serviceWatch() {
	for {
		res, err := m.watcher.Next()
		if err != nil {
			logger.Infof("ServiceWatch error %s", err.Error())
			if err.Error() == "could not get next" {
				return
			}
		}
		switch res.Action {
		case "delete":
			for _, node := range res.Service.Nodes {
				m.nodeChannel <- &nodeChangeOpt{Opt: NodeOptDelete, Address: node.Address}
			}
		case "update", "create":
			for _, node := range res.Service.Nodes {
				m.nodeChannel <- &nodeChangeOpt{Opt: NodeOptUpdate, Address: node.Address}
			}
		}
	}
}

func (m *defaultMgr) initWatch() {
	watcher, err := registry.DefaultRegistry.Watch(registry.WatchService("match_evaluator"), registry.WatchDomain(os.Getenv("MICRO_NAMESPACE")))
	if err != nil {
		logger.Fatal("initwatch have error %s", err.Error())
		return
	}
	if m.watcher != nil {
		m.watcher.Stop()
		m.watcher = nil
	}
	m.watcher = watcher
	svrList, errs := registry.DefaultRegistry.GetService("match_evaluator", registry.GetDomain(os.Getenv("MICRO_NAMESPACE")))
	if errs != nil {
		logger.Infof("initwatch GetService error %s", errs.Error())
	}
	for _, srv := range svrList {
		for _, node := range srv.Nodes {
			m.nodeChannel <- &nodeChangeOpt{Opt: NodeOptUpdate, Address: node.Address}
		}
	}
}

func (m *defaultMgr) Start() error {
	//consistenthash.New(replicas int, fn Hash)
	m.initWatch()
	go m.serviceWatch()
	go m.loop()
	return nil
}

func (m *defaultMgr) Stop() error {
	close(m.exited)
	return nil
}

func (m *defaultMgr) AddEvalOpt(req *match_evaluator.ToEvalReq, key string) {
	m.evalChannel <- &evalOpt{Req: req, Key: key}
}

func (m *defaultMgr) loop() {
	for {
		select {
		case <-m.exited:
			return
		case evalOpt := <-m.evalChannel:
			addr := m.ring.Get(evalOpt.Key)
			go m.sendEvalReq(evalOpt.Req, addr)
		case opt := <-m.nodeChannel:
			if opt.Opt == NodeOptUpdate {
				m.ring.Add(opt.Address)
			} else if opt.Opt == NodeOptDelete {
				m.ring.Remove(opt.Address)
			}
		}
	}
}

func (m *defaultMgr) sendEvalReq(req *match_evaluator.ToEvalReq, addr string) {
	evalSrv := match_evaluator.NewMatchEvaluatorService("match_evaluator", client.DefaultClient)
	_, err := evalSrv.ToEval(context.Background(), req, client.WithAddress(addr))
	if err != nil {
		logger.Infof("ToEval error %+v", err)
	} else {
		//logger.Infof("ToEval result %+v", evalRsp)
	}
}

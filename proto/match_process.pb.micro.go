// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: proto/match_process.proto

package match_process

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "context"
	api "github.com/micro/micro/v3/service/api"
	client "github.com/micro/micro/v3/service/client"
	server "github.com/micro/micro/v3/service/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ api.Endpoint
var _ context.Context
var _ client.Option
var _ server.Option

// Api Endpoints for MatchProcess service

func NewMatchProcessEndpoints() []*api.Endpoint {
	return []*api.Endpoint{}
}

// Client API for MatchProcess service

type MatchProcessService interface {
	MatchTask(ctx context.Context, in *MatchTaskReq, opts ...client.CallOption) (*MatchTaskRsp, error)
}

type matchProcessService struct {
	c    client.Client
	name string
}

func NewMatchProcessService(name string, c client.Client) MatchProcessService {
	return &matchProcessService{
		c:    c,
		name: name,
	}
}

func (c *matchProcessService) MatchTask(ctx context.Context, in *MatchTaskReq, opts ...client.CallOption) (*MatchTaskRsp, error) {
	req := c.c.NewRequest(c.name, "MatchProcess.MatchTask", in)
	out := new(MatchTaskRsp)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for MatchProcess service

type MatchProcessHandler interface {
	MatchTask(context.Context, *MatchTaskReq, *MatchTaskRsp) error
}

func RegisterMatchProcessHandler(s server.Server, hdlr MatchProcessHandler, opts ...server.HandlerOption) error {
	type matchProcess interface {
		MatchTask(ctx context.Context, in *MatchTaskReq, out *MatchTaskRsp) error
	}
	type MatchProcess struct {
		matchProcess
	}
	h := &matchProcessHandler{hdlr}
	return s.Handle(s.NewHandler(&MatchProcess{h}, opts...))
}

type matchProcessHandler struct {
	MatchProcessHandler
}

func (h *matchProcessHandler) MatchTask(ctx context.Context, in *MatchTaskReq, out *MatchTaskRsp) error {
	return h.MatchProcessHandler.MatchTask(ctx, in, out)
}

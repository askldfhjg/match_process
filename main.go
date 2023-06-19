package main

import (
	"match_process/handler"
	pb "match_process/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	srv := service.New(
		service.Name("match_process"),
		service.Version("latest"),
	)

	// Register handler
	pb.RegisterMatchProcessHandler(srv.Server(), new(handler.Match_process))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

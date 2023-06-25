package main

import (
	"match_process/handler"
	"match_process/internal/db"
	"match_process/internal/db/redis"
	"match_process/internal/manager"
	"match_process/process"
	pb "match_process/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	process.DefaultManager = manager.NewManager()
	srv := service.New(
		service.Name("match_process"),
		service.Version("latest"),
		service.BeforeStart(func() error {
			svr, err := redis.New(
				db.WithAddress("127.0.0.1:6379"),
				db.WithPoolMaxActive(5),
				db.WithPoolMaxIdle(100),
				db.WithPoolIdleTimeout(300))
			if err != nil {
				return err
			}
			db.Default = svr
			return nil
		}),
		service.AfterStart(process.DefaultManager.Start),
		service.BeforeStop(process.DefaultManager.Stop),
	)

	// Register handler
	pb.RegisterMatchProcessHandler(srv.Server(), new(handler.Match_process))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

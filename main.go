package main

import (
	"context"
	"ctacampado/go-chi-service/pkg/logger"
)

const svcName = "go-chi-service"
const dev = "localhost:8080"
const ver = "v0.0.0"

func main() {
	logger.Init(svcName, ver)
	log := logger.Default()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svr, err := NewServer(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	if err := svr.Run(dev); err != nil {
		log.Fatal().Err(err).Msg("Cannot run server")
	}
}

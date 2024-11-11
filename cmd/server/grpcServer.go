package main

import (
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/dglazkoff/go-metrics/cmd/server/config"
	"github.com/dglazkoff/go-metrics/cmd/server/router"
	"github.com/dglazkoff/go-metrics/cmd/server/services/service"
	"github.com/dglazkoff/go-metrics/cmd/server/storage"
	"google.golang.org/grpc"

	pb "github.com/dglazkoff/go-metrics/internal/models/proto"
)

func RunGRPCServer(cfg *config.Config, errChan chan<- error) *grpc.Server {
	store, fileStorage, err := storage.InitStorages(cfg)

	if err != nil {
		errChan <- err
		return nil
	}

	// logger.Log.Infow("Starting gRPC Server on ", "addr", cfg.RunAddr)

	if cfg.StoreInterval != 0 {
		go fileStorage.WriteMetrics(true)
	}

	metricService := service.New(store, fileStorage, cfg)

	listen, err := net.Listen("tcp", cfg.RunAddr)
	if err != nil {
		errChan <- err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServer(grpcServer, router.NewMetricsServer(metricService))

	go func() {
		if err := grpcServer.Serve(listen); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	return grpcServer
}

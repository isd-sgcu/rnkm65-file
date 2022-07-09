package main

import (
	"context"
	"fmt"
	fRepo "github.com/isd-sgcu/rnkm65-file/src/app/repository/file"
	gcsSrv "github.com/isd-sgcu/rnkm65-file/src/app/service/gcs"
	gcsClt "github.com/isd-sgcu/rnkm65-file/src/client/gcs"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/isd-sgcu/rnkm65-file/src/database"
	"github.com/isd-sgcu/rnkm65-file/src/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type operation func(ctx context.Context) error

func gracefulShutdown(ctx context.Context, timeout time.Duration, ops map[string]operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-s

		log.Info().
			Str("service", "graceful shutdown").
			Msgf("got signal \"%v\" shutting down service", sig)

		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Error().
				Str("service", "graceful shutdown").
				Msgf("timeout %v ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		for key, op := range ops {
			wg.Add(1)
			innerOp := op
			innerKey := key
			go func() {
				defer wg.Done()

				log.Info().
					Str("service", "graceful shutdown").
					Msgf("cleaning up: %v", innerKey)
				if err := innerOp(ctx); err != nil {
					log.Error().
						Str("service", "graceful shutdown").
						Err(err).
						Msgf("%v: clean up failed: %v", innerKey, err.Error())
					return
				}

				log.Info().
					Str("service", "graceful shutdown").
					Msgf("%v was shutdown gracefully", innerKey)
			}()
		}

		wg.Wait()
		close(wait)
	}()

	return wait
}

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "file").
			Msg("Failed to start service")
	}

	db, err := database.InitDatabase(&conf.Database)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "file").
			Msg("Failed to start service")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", conf.App.Port))
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "file").
			Msg("Failed to start service")
	}

	fileRepo := fRepo.NewRepository(db)

	gcsClient := gcsClt.NewClient(conf.GCS)
	fileSrv := gcsSrv.NewService(conf.GCS, gcsClient, fileRepo)

	grpcServer := grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	proto.RegisterFileServiceServer(grpcServer, fileSrv)

	reflection.Register(grpcServer)
	go func() {
		log.Info().
			Str("service", "file").
			Msgf("RNKM65 file starting at port %v", conf.App.Port)

		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal().
				Err(err).
				Str("service", "auth").
				Msg("Failed to start service")
		}
	}()

	wait := gracefulShutdown(context.Background(), 2*time.Second, map[string]operation{
		"server": func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
	})

	<-wait

	grpcServer.GracefulStop()
	log.Info().
		Str("service", "file").
		Msg("Closing the listener")
	lis.Close()
	log.Info().
		Str("service", "file").
		Msg("End of Program")
}

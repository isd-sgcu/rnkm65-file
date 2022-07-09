package main

import (
	"context"
	"fmt"
	gcsSrv "github.com/isd-sgcu/rnkm65-file/src/app/service/gcs"
	gcsClt "github.com/isd-sgcu/rnkm65-file/src/client/gcs"
	"github.com/isd-sgcu/rnkm65-file/src/config"
	"github.com/isd-sgcu/rnkm65-file/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log"
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

		log.Printf("got signal \"%v\" shutting down service", sig)

		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Printf("timeout %v ms has been elapsed, force exit", timeout.Milliseconds())
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

				log.Printf("cleaning up: %v", innerKey)
				if err := innerOp(ctx); err != nil {
					log.Printf("%v: clean up failed: %v", innerKey, err.Error())
					return
				}

				log.Printf("%v was shutdown gracefully", innerKey)
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
		log.Fatal("Cannot load config", err.Error())
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", conf.App.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gcsClient := gcsClt.NewClient(conf.GCS)
	fileSrv := gcsSrv.NewService(conf.GCS, gcsClient)

	grpcServer := grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	proto.RegisterFileServiceServer(grpcServer, fileSrv)

	reflection.Register(grpcServer)
	go func() {
		log.Println(fmt.Sprintf("RNKM65 backend starting at port %v", conf.App.Port))

		if err = grpcServer.Serve(lis); err != nil {
			log.Fatalln("Failed to serve:", err)
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
	log.Println("Closing the listener")
	lis.Close()
	log.Println("End of Program")
}

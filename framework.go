package framework

import (
	"context"
	"fmt"
	"github.com/RealJonathanYip/framework/config"
	"github.com/RealJonathanYip/framework/interceptor"
	"github.com/RealJonathanYip/framework/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

var (
	listener net.Listener
	port     uint16
	server   *grpc.Server
)

const (
	START_PORT = 6000
)

func init() {

	output := config.FrameWorkConfig.LogOutput
	log.InitLog(log.SetTarget(output.Value), log.LogFilePath(output.Path), log.LogFileRotate(output.FileRotate))
	log.SetLogLevel(config.FrameWorkConfig.LogLevel)

	server = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		interceptor.WithServerTraceInterceptor(),
	)))

	tryCount := 1000
	for i := 0; i < tryCount; i++ {
		port = START_PORT + uint16(i)
		listenerTemp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}

		listener = listenerTemp
		log.Infof(context.TODO(), "server listen at:%d", port)
		return
	}

	log.Panic(context.TODO(), "start listener fail!")
}

func GetGrpcServer() *grpc.Server {
	return server
}

func Serve() {
	//TODO: add service discover logic...
	if err := server.Serve(listener); err != nil {
		log.Panicf(context.TODO(), "failed to serve: %v", err)
	}
}

func GetGrpcConnection(serviceName string) (*grpc.ClientConn, error) {
	//TODO: add service discover logic...
	address := "127.0.0.1:8088"
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), interceptor.WithClientUnaryInterceptor())
	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}

		log.Fatalf(context.TODO(), "connect to %s-%v fail: %v", serviceName, address, err)
		return nil, err
	}

	return conn, nil
}

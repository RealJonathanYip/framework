package rpc_server

import (
	"context"
	"fmt"
	"github.com/RealJonathanYip/framework/interceptor"
	"github.com/RealJonathanYip/framework/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

type RpcServer struct {
	listener net.Listener
	port     uint16
	server   *grpc.Server
	name     string
}

func New(name string) *RpcServer {
	return &RpcServer{
		name: "rpc." + name,
		server: grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			interceptor.WithServerTraceInterceptor(),
		))),
	}
}

func (r *RpcServer) Serve() {
	//TODO: add service discover logic...
	startPort, tryCount := 8888, 1000
	for i := 0; i < tryCount; i++ {
		port := startPort + i
		listenerTemp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Debugf(context.TODO(), "server:%v listen at:%d fail:%v", r.name, port, err)
			continue
		}

		r.listener = listenerTemp
		log.Infof(context.TODO(), "server:%v listen at:%d", r.name, port)

		if err := r.server.Serve(r.listener); err != nil {
			log.Panicf(context.TODO(), "failed to serve: %v", err)
		}
	}
}

func GetRpcServiceConnection(serviceName string) (*grpc.ClientConn, error) {
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

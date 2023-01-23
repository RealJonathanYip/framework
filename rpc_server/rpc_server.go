package rpc_server

import (
	"context"
	"fmt"
	context0 "github.com/RealJonathanYip/framework/context0"
	"github.com/RealJonathanYip/framework/interceptor"
	"github.com/RealJonathanYip/framework/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
	"time"
)

type RpcServer struct {
	listener net.Listener
	port     uint16
	server   *grpc.Server
	name     string
}

func New(name string) *RpcServer {
	rpcServer := &RpcServer{name: "rpc." + name}
	rpcServer.server = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		rpcServer.WithServerTraceInterceptor(),
	)))

	return rpcServer
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

func (r *RpcServer) GetRpcServiceConnection(serviceName string) (*grpc.ClientConn, error) {
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

// TODO：log and ip trace
func (r *RpcServer) WithServerTraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = context0.FromRpcContext(ctx)

		upstreamService, exist := context0.Get(ctx, context0.ContextKeyUpstreamService)
		if !exist {
			upstreamService = "unknow"
		}

		upstreamMethod, exist := context0.Get(ctx, context0.ContextKeyUpstreamMethod)
		if !exist {
			upstreamMethod = "unknow"
		}

		methodInfos := strings.Split(info.FullMethod, "/")
		method := info.FullMethod
		service := r.name
		if len(methodInfos) == 3 {
			method = methodInfos[2]
			service = fmt.Sprintf("<grpc>-<%s>", methodInfos[1])
		}

		var upstreamAddress string
		if peerTemp, ok := peer.FromContext(ctx); ok {
			if tcpAddr, ok := peerTemp.Addr.(*net.TCPAddr); ok {
				upstreamAddress = tcpAddr.String()
			} else {
				upstreamAddress = peerTemp.Addr.String()
			}
		}
		context0.Set(ctx, context0.ContextKeyCurrentMethod, method,
			context0.ContextKeyCurrentService, service,
			context0.ContextKeyUpstreamAddress, upstreamAddress)

		now := time.Now()
		resp, err := handler(ctx, req)
		cost := time.Since(now).Milliseconds()

		log.Infof(ctx, "【serve】upstreamAddress:%s upstreamService:%v upstreamMethod:%v service:%v method:%v cost:%v(ms) req:%+v, resp:%+v",
			upstreamAddress, upstreamService, upstreamMethod, service, method, cost, req, resp)

		return resp, err
	}
}

package interceptor

import (
	"context"
	"fmt"
	"github.com/RealJonathanYip/framework/log"
	"github.com/RealJonathanYip/framework/rpc_context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	processName string
)

func init() {
	processName = fmt.Sprintf("<process>-<%s>", path.Base(os.Args[0]))
}

// TODO：log and ip trace
func WithServerTraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = rpc_context.NewContext(ctx)

		upstreamService, exist := rpc_context.Get(ctx, rpc_context.ContextKeyUpstreamService)
		if !exist {
			upstreamService = "unknow"
		}

		upstreamMethod, exist := rpc_context.Get(ctx, rpc_context.ContextKeyUpstreamMethod)
		if !exist {
			upstreamMethod = "unknow"
		}

		methodInfos := strings.Split(info.FullMethod, "/")
		method := info.FullMethod
		service := processName
		if len(methodInfos) == 3 {
			method = methodInfos[2]
			service = fmt.Sprintf("<grpc>-<%s>", methodInfos[1])
		}
		rpc_context.Set(ctx, rpc_context.ContextKeyCurrentMethod, method)
		rpc_context.Set(ctx, rpc_context.ContextKeyCurrentService, service)

		var upstreamAddress string
		if peer, ok := peer.FromContext(ctx); ok {
			if tcpAddr, ok := peer.Addr.(*net.TCPAddr); ok {
				upstreamAddress = tcpAddr.String()
			} else {
				upstreamAddress = peer.Addr.String()
			}
		}
		rpc_context.Set(ctx, rpc_context.ContextKeyUpstreamAddress, upstreamAddress)

		now := time.Now()
		resp, err := handler(ctx, req)
		cost := time.Since(now).Milliseconds()

		log.Infof(ctx, "【serve】upstreamAddress:%s upstreamService:%v upstreamMethod:%v service:%v method:%v cost:%v(ms) req:%+v, resp:%+v",
			upstreamAddress, upstreamService, upstreamMethod, service, method, cost, req, resp)

		return resp, err
	}
}

func WithClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(
		ctx context.Context,
		method string,
		req interface{},
		resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = rpc_context.Copy(ctx)

		upstreamService, exist := rpc_context.Get(ctx, rpc_context.ContextKeyUpstreamService)
		if !exist {
			upstreamService = "unknow"
		}
		upstreamMethod, exist := rpc_context.Get(ctx, rpc_context.ContextKeyUpstreamMethod)
		if !exist {
			upstreamMethod = "unknow"
		}

		currentMethod, exist := rpc_context.Get(ctx, rpc_context.ContextKeyCurrentMethod)
		if !exist {
			pc := make([]uintptr, 1)
			runtime.Callers(4, pc)
			function := runtime.FuncForPC(pc[0])
			currentMethod = fmt.Sprintf("<local>-<%s>", function.Name())
		}
		rpc_context.Set(ctx, rpc_context.ContextKeyUpstreamMethod, currentMethod)

		currentService, exist := rpc_context.Get(ctx, rpc_context.ContextKeyCurrentService)
		if !exist {
			currentService = processName
		}
		rpc_context.Set(ctx, rpc_context.ContextKeyUpstreamService, currentService)

		methodInfos := strings.Split(method, "/")
		downstreamMethod := method
		downstreamService := "unknow"
		if len(methodInfos) == 3 {
			downstreamMethod = methodInfos[2]
			downstreamService = methodInfos[1]
		}

		rpc_context.Del(ctx, rpc_context.ContextKeyUpstreamAddress)

		now := time.Now()
		err := invoker(rpc_context.NewRpcContext(ctx), method, req, resp, cc, opts...)
		cost := time.Since(now).Milliseconds()

		log.Infof(ctx, "【request】upstreamService:%v upstreamMethod:%v downstreamService:%v downstreamMethod:%v currentService:%v currentMethod:%v cost:%v(ms) req:%+v, resp:%+v \n",
			upstreamService, upstreamMethod, downstreamService, downstreamMethod, currentService, currentMethod, cost, req, resp)

		return err
	})
}

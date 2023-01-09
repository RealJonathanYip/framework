package framework

//context interface for business layer

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"sync"
)

const (
	ContextKeyTraceID         = "persist_trace_id"
	ContextKeyUpstreamService = "temp_upstream_service"
	ContextKeyUpstreamMethod  = "temp_upstream_method"
	ContextKeyCurrentMethod   = "temp_current_method"
	ContextKeyCurrentService  = "temp_current_service"
	ContextKeyUpstreamAddress = "temp_upstream_address"
	contextMeta               = "meta_data"
)

type metaDataInner struct {
	metaData metadata.MD
	sync.RWMutex
}

func Get(ctx context.Context, key string) (string, bool) {
	meta, ok := ctx.Value(contextMeta).(*metaDataInner)
	if !ok {
		return "", false
	}

	meta.RLock()
	defer meta.RUnlock()
	values := meta.metaData.Get(key)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

func Set(ctx context.Context, kvs ...string) bool {
	meta, ok := ctx.Value(contextMeta).(*metaDataInner)
	if !ok {
		fmt.Println("ctx is not rpc_context, please check")
		return false
	}

	meta.Lock()
	defer meta.Unlock()
	if len(kvs)%2 == 1 {
		panic(fmt.Sprintf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kvs)))
	}

	for i := 0; i < len(kvs); i += 2 {
		meta.metaData.Set(kvs[i], kvs[i+1])
	}

	return true
}

func Del(ctx context.Context, keys ...string) {
	meta, ok := ctx.Value(contextMeta).(*metaDataInner)
	if !ok {
		return
	}

	meta.Lock()
	defer meta.Unlock()
	for _, key := range keys {
		meta.metaData.Delete(key)
	}
}

func NewContext(ctx context.Context) context.Context {
	meta, exit := metadata.FromIncomingContext(ctx)
	if !exit {
		return context.WithValue(ctx, contextMeta, &metaDataInner{metaData: metadata.Pairs(ContextKeyTraceID, uuid.New().String())})
	}

	return context.WithValue(ctx, contextMeta, &metaDataInner{metaData: meta})
}

func Copy(from context.Context) context.Context {
	meta, ok := from.Value(contextMeta).(*metaDataInner)
	if !ok {
		return context.WithValue(from, contextMeta, &metaDataInner{metaData: metadata.Pairs(ContextKeyTraceID, uuid.New().String())})
	}

	meta.RLock()
	defer meta.RUnlock()
	return context.WithValue(from, contextMeta, &metaDataInner{metaData: meta.metaData.Copy()})
}

func NewRpcContext(ctx context.Context) context.Context {
	meta := ctx.Value(contextMeta).(*metaDataInner)
	if meta == nil {
		return metadata.NewOutgoingContext(ctx, metadata.Pairs(ContextKeyTraceID, uuid.New().String()))
	}

	meta.RLock()
	defer meta.RUnlock()
	return metadata.NewOutgoingContext(ctx, meta.metaData)
}

func GetLogText(ctx context.Context) string {
	traceID, exist := Get(ctx, ContextKeyTraceID)
	if !exist {
		traceID = "unknow"
	}

	method, exist := Get(ctx, ContextKeyCurrentMethod)
	if !exist {
		method = "unknow"
	}

	service, exist := Get(ctx, ContextKeyCurrentService)
	if !exist {
		service = "unknow"
	}

	return fmt.Sprintf(" traceID:%s service:%s method:%s", traceID, service, method)
}

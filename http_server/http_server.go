package http_server

import (
	"context"
	"fmt"
	"github.com/RealJonathanYip/framework/context0"
	"github.com/RealJonathanYip/framework/log"
	"github.com/RealJonathanYip/framework/overflow"
	"github.com/RealJonathanYip/framework/utils"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

type HttpServer struct {
	httpRouter      map[string]func(context.Context, *Response, *Request)
	onBeforeRequest []func(context.Context, *Response, *Request) bool
	onBeforeReply   []func(context.Context, *Response, *Request)
	listener        net.Listener
	port            int
	name            string
}

const (
	ERROR_SERVICE_NOT_AVAILABLE = 502
	ERROR_AUTH_ERROR            = 401
	_METHOD_POST                = "POST"
	_METHOD_GET                 = "GET"
	_METHOD_DELETE              = "DELETE"
	_METHOD_PUT                 = "PUT"
)

// 公用的返回
type Reply struct {
	Result uint32      `json:"result"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func New(name string) *HttpServer {
	return &HttpServer{
		httpRouter:      make(map[string]func(context.Context, *Response, *Request)),
		onBeforeRequest: make([]func(context.Context, *Response, *Request) bool, 0),
		onBeforeReply:   make([]func(context.Context, *Response, *Request), 0),
		name:            "web." + name,
	}
}

func (h *HttpServer) onReq(rsp http.ResponseWriter, req *http.Request) {
	var path = req.URL.Path
	var method = req.Method

	if method == "OPTIONS" {
		rsp.WriteHeader(200)
		return
	}
	ctx := context0.NewContext()
	defer utils.Recover(ctx)

	szEntryPoint := path + "_" + method
	context0.Set(ctx, context0.ContextKeyCurrentService, h.name, context0.ContextKeyCurrentMethod, szEntryPoint)

	if fnHandler, bExist := h.httpRouter[szEntryPoint]; !bExist {
		log.Warningf(ctx, "not found http -> %v", path+"_"+method)
		http.NotFound(rsp, req)
		return
	} else {
		request := &Request{req}
		onBeforeReply := func(ctx context.Context, response *Response) {
			for _, handler := range h.onBeforeReply {
				handler(ctx, response, request)
			}
		}
		response := &Response{rsp, onBeforeReply}

		fnHandler(ctx, response, request)
	}
}

func (h *HttpServer) doRegisterHttpHandler(path, method string, handler, overFlowHandler func(context.Context, *Response, *Request), maxQPS ...uint32) {
	qps := uint32(10240)
	if len(maxQPS) > 0 {
		qps = maxQPS[0]
	}

	h.httpRouter[path+"_"+method] = func(ctx context.Context, resp *Response, req *Request) {
		if overflow.IsOverFlow(method+"."+path, qps) {
			if overFlowHandler != nil {
				overFlowHandler(ctx, resp, req)
				return
			}

			http.Error(resp, "uri over flow!  plase try again later", ERROR_SERVICE_NOT_AVAILABLE)
			return
		}

		for _, fnHandler := range h.onBeforeRequest {
			if exit := fnHandler(ctx, resp, req); exit {
				return
			}
		}

		handler(ctx, resp, req)
	}

	log.Infof(context0.NewContext(), "register http router : %v", path+"_"+method)
}

func (h *HttpServer) Post(szPath string, fnHandler, fnOnOverFlow func(context.Context, *Response, *Request), maxQPS ...uint32) {
	if _, bExist := h.httpRouter[_METHOD_POST+"."+szPath]; !bExist {
		h.doRegisterHttpHandler(szPath, _METHOD_POST, fnHandler, fnOnOverFlow, maxQPS...)
	} else {
		log.Panicf(context0.NewContext(), "http uri:%s exist!", _METHOD_POST+"."+szPath)
	}
}

func (h *HttpServer) Put(szPath string, fnHandler, fnOnOverFlow func(context.Context, *Response, *Request), maxQPS ...uint32) {
	if _, bExist := h.httpRouter[_METHOD_PUT+"."+szPath]; !bExist {
		h.doRegisterHttpHandler(szPath, _METHOD_PUT, fnHandler, fnOnOverFlow, maxQPS...)
	} else {
		log.Panicf(context0.NewContext(), "http uri:%s exist!", _METHOD_PUT+"."+szPath)
	}
}

func (h *HttpServer) Get(szPath string, fnHandler, fnOnOverFlow func(context.Context, *Response, *Request), maxQPS ...uint32) {
	if _, bExist := h.httpRouter[_METHOD_GET+"."+szPath]; !bExist {
		h.doRegisterHttpHandler(szPath, _METHOD_GET, fnHandler, fnOnOverFlow, maxQPS...)
	} else {
		log.Panicf(context0.NewContext(), "http uri:%s exist!", _METHOD_GET+"."+szPath)
	}
}

func (h *HttpServer) Delete(szPath string, fnHandler, fnOnOverFlow func(context.Context, *Response, *Request), maxQPS ...uint32) {
	if _, bExist := h.httpRouter[_METHOD_DELETE+"."+szPath]; !bExist {
		h.doRegisterHttpHandler(szPath, _METHOD_DELETE, fnHandler, fnOnOverFlow, maxQPS...)
	} else {
		log.Panicf(context0.NewContext(), "http uri:%s exist!", _METHOD_DELETE+"."+szPath)
	}
}

func (h *HttpServer) Run() error {
	startPort, tryCount := 6666, 1000

	for i := 0; i < tryCount; i++ {
		port := startPort + i
		listenerTemp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Warningf(context.TODO(), "start http server:%s listener fail!:%v", h.name, err)
			continue
		}

		h.port = port
		h.listener = listenerTemp
		log.Infof(context.TODO(), "http server:%v listen at:%d", h.name, port)

		mux := http.NewServeMux()
		mux.HandleFunc("/", h.onReq)

		//TODO: add service discover logic

		err = http.Serve(h.listener, mux)
		if err != nil {
			log.Warningf(context.TODO(), "start http server:%s fail!:%v", h.name, err)
			continue
		}

		return nil
	}

	return errors.Errorf("http server:%s fail too much", h.name)
}

func (h *HttpServer) OnBeforeRequest(handler func(context.Context, *Response, *Request) bool) {
	h.onBeforeRequest = append(h.onBeforeRequest, handler)
}

func (h *HttpServer) OnBeforeReply(handler func(context.Context, *Response, *Request)) {
	h.onBeforeReply = append(h.onBeforeReply, handler)
}

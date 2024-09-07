package http_server

import (
	"context"
	"encoding/json"
	"github.com/RealJonathanYip/framework/log"
	"net/http"
)

type Response struct {
	http.ResponseWriter
	onBeforeReply func(context.Context, *Response)
}

func (r *Response) ReplyJson(ctx context.Context, data interface{}) error {
	byteData, ok := data.([]byte)
	if !ok {
		byteDataTemp, err := json.Marshal(data)
		if err != nil {
			log.Warningf(ctx, "json marshal result err!:%v", err)
			return err
		}

		byteData = byteDataTemp
	}

	r.Header().Set("content-type", "application/json;utf-8")
	r.onBeforeReply(ctx, r)

	if _, err := r.Write(byteData); err != nil {
		log.Warningf(ctx, "http reply err!:%v", err)
		return err
	}

	return nil
}

package http_server

import (
	"context"
	"encoding/json"
	"github.com/RealJonathanYip/framework/log"
	"net/http"
	"net/url"
)

type QueryGetter struct {
	query url.Values
}

type HeaderGetter struct {
	header http.Header
}

type FormGetter struct {
	form url.Values
}

func NewQueryGetter(req *http.Request) *QueryGetter {
	return &QueryGetter{
		query: req.URL.Query(),
	}
}

func (q *QueryGetter) Get(name string) (string, bool) {
	if !q.query.Has(name) {
		return "", false
	}

	return q.query.Get(name), true
}

func NewHeaderGetter(req *http.Request) *HeaderGetter {
	return &HeaderGetter{
		header: req.Header,
	}
}

func (q *HeaderGetter) Get(name string) (string, bool) {
	values := q.header.Values(name)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

func NewFormGetter(req *http.Request) *FormGetter {
	_ = req.ParseForm()
	return &FormGetter{
		form: req.PostForm,
	}
}

func (q *FormGetter) Get(name string) (string, bool) {
	if !q.form.Has(name) {
		return "", false
	}

	return q.form.Get(name), true
}

func ReplyAny(ctx context.Context, resp *http.ResponseWriter, data interface{}) error {
	byteData, ok := data.([]byte)
	if !ok {
		byteDataTemp, err := json.Marshal(data)
		if err != nil {
			log.Warningf(ctx, "json marshal result err!:%v", err)
			return err
		}

		byteData = byteDataTemp
	}

	(*resp).Header().Set("content-type", "application/json;utf-8")

	if _, err := (*resp).Write(byteData); err != nil {
		log.Warningf(ctx, "http reply err!:%v", err)
		return err
	}

	return nil
}

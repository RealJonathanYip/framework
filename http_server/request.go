package http_server

import (
	"github.com/bytedance/go-tagexpr/v2/binding"
	"net/http"
	"net/url"
)

type queryGetter struct {
	query url.Values
}

type headerGetter struct {
	header http.Header
}

type formGetter struct {
	form url.Values
}

type Request struct {
	*http.Request
}

func (r *Request) ParamsFromQuery(params interface{}) error {
	err := binding.BindAndValidate(params, r.Request, newQueryGetter(r.Request))
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) ParamsFromHeader(params interface{}) error {
	err := binding.BindAndValidate(params, r.Request, newHeaderGetter(r.Request))
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) ParamsFromForm(params interface{}) error {
	err := binding.BindAndValidate(params, r.Request, newFormGetter(r.Request))
	if err != nil {
		return err
	}

	return nil
}

func newQueryGetter(req *http.Request) *queryGetter {
	return &queryGetter{
		query: req.URL.Query(),
	}
}

func (q *queryGetter) Get(name string) (string, bool) {
	if !q.query.Has(name) {
		return "", false
	}

	return q.query.Get(name), true
}

func newHeaderGetter(req *http.Request) *headerGetter {
	return &headerGetter{
		header: req.Header,
	}
}

func (q *headerGetter) Get(name string) (string, bool) {
	values := q.header.Values(name)
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

func newFormGetter(req *http.Request) *formGetter {
	_ = req.ParseForm()
	return &formGetter{
		form: req.PostForm,
	}
}

func (q *formGetter) Get(name string) (string, bool) {
	if !q.form.Has(name) {
		return "", false
	}

	return q.form.Get(name), true
}

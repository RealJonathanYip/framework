package http_server

import (
	"github.com/bytedance/go-tagexpr/v2/binding"
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

type Request struct {
	http.Request
}

func (r *Request) ParamsFromQuery(params *struct{}) error {
	err := binding.BindAndValidate(&params, r, newQueryGetter(&r.Request))
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) ParamsFromHeader(params *struct{}) error {
	err := binding.BindAndValidate(&params, r, newHeaderGetter(&r.Request))
	if err != nil {
		return err
	}

	return nil
}

func (r *Request) ParamsFromForm(params *struct{}) error {
	err := binding.BindAndValidate(&params, r, newFormGetter(&r.Request))
	if err != nil {
		return err
	}

	return nil
}

func newQueryGetter(req *http.Request) *QueryGetter {
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

func newHeaderGetter(req *http.Request) *HeaderGetter {
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

func newFormGetter(req *http.Request) *FormGetter {
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

package jeen

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Request struct {
	instance *http.Request
}

// create new request instance
func newRequest(r *http.Request) *Request {
	return &Request{
		instance: r,
	}
}

// return instance of *http.Request
func (r *Request) Instance() *http.Request {
	return r.instance
}

func (r *Request) IsTLS() bool {
	return r.instance.TLS != nil
}

func (r *Request) RemoteAddr() string {
	return r.instance.RemoteAddr
}

func (r *Request) RequestURI() string {
	return r.instance.RequestURI
}

func (r *Request) URLParam(key string) string {
	return chi.URLParam(r.instance, key)
}

func (r *Request) QueryParam(name string) string {
	return r.instance.URL.Query().Get(name)
}

func (r *Request) QueryString() string {
	return r.instance.URL.RawQuery
}

func (r *Request) IsWebSocket() bool {
	upgrade := r.instance.Header.Get("Upgrade")
	return strings.EqualFold(upgrade, "websocket")
}

func (r *Request) Scheme() string {
	if r.IsTLS() {
		return "https"
	}
	if scheme := r.instance.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if scheme := r.instance.Header.Get("X-Forwarded-Protocol"); scheme != "" {
		return scheme
	}
	if ssl := r.instance.Header.Get("X-Forwarded-Ssl"); ssl == "on" {
		return "https"
	}
	if scheme := r.instance.Header.Get("X-Url-Scheme"); scheme != "" {
		return scheme
	}
	return "http"
}

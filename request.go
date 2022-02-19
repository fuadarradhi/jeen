package jeen

import (
	"net/http"
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

// func (r *Request) IsTLS() bool {
// 	return r.request.TLS != nil
// }

// func (r *Request) RealIP() string {
// 	return r.request.RemoteAddr
// }

// // URLParam returns the url parameter from a http.Request object.
// func (r *Request) URLParam(key string) string {
// 	return chi.URLParam(r.request, key)
// }

// func (r *Request) FormValue(name string) string {
// 	return r.request
// }

// func (r *Request) QueryString() string {
// 	return r.request.URL.RawQuery
// }

// func (r *Request) QueryParam(name string) string {
// 	return r.request.URL.Query().Get(name)
// }

// func (r *Request) IsWebSocket() bool {
// 	upgrade := r.request.Header.Get("Upgrade")
// 	return strings.EqualFold(upgrade, "websocket")
// }

// func (r *Request) FormParams() (url.Values, error) {
// 	if strings.HasPrefix(r.request.Header.Get("Content-Type"), "multipart/form-data") {
// 		if err := r.request.ParseMultipartForm(32 << 20); err != nil {
// 			return nil, err
// 		}
// 	} else {
// 		if err := r.request.ParseForm(); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return r.request.Form, nil
// }

// func (r *Request) FormFile(name string) (*multipart.FileHeader, error) {
// 	f, fh, err := r.request.FormFile(name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	f.Close()
// 	return fh, nil
// }

// func (r *Request) MultipartForm() (*multipart.Form, error) {
// 	err := r.request.ParseMultipartForm(32 << 20)
// 	return r.request.MultipartForm, err
// }

// func (r *Request) Scheme() string {
// 	// Can't use `r.Request.URL.Scheme`
// 	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
// 	if r.IsTLS() {
// 		return "https"
// 	}
// 	if scheme := r.request.Header.Get("X-Forwarded-Proto"); scheme != "" {
// 		return scheme
// 	}
// 	if scheme := r.request.Header.Get("X-Forwarded-Protocol"); scheme != "" {
// 		return scheme
// 	}
// 	if ssl := r.request.Header.Get("X-Forwarded-Ssl"); ssl == "on" {
// 		return "https"
// 	}
// 	if scheme := r.request.Header.Get("X-Url-Scheme"); scheme != "" {
// 		return scheme
// 	}
// 	return "http"
// }

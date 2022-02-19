package jeen

import (
	"context"
	"errors"
	"net/http"
)

// All resource needed for development
type Resource struct {
	// private for internal use
	writer  http.ResponseWriter
	request *http.Request

	// request
	Request *Request

	// writer
	Writer *Writer

	// cookie
	Cookie *Cookie

	// request context
	Context context.Context

	// database sql.DB
	Database *Database

	// session SCS
	Session *Session

	// html response with goview
	Html *Html

	// json response
	Json *Json
}

// create new resource
func createResource(rw http.ResponseWriter, r *http.Request, h *HtmlEngine) *Resource {
	return &Resource{
		// private
		request: r,
		writer:  rw,

		Context: r.Context(),
		Request: newRequest(r),
		Writer:  newWriter(rw),
		Cookie:  newCookie(rw, r),
		Html:    newHtml(rw, h),
		Json:    newJson(rw),
	}
}

// Set set value to request context, see Get to get it
func (r *Resource) Set(key, val interface{}) {
	r.request = r.request.WithContext(context.WithValue(r.Context, key, val))
}

// Get get value from context
func (r *Resource) Get(key interface{}) interface{} {
	return r.Context.Value(key)
}

// Redirect redirects the request to a provided URL with status code.
func (r *Resource) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return errors.New("invalid redirect status code")
	}
	http.Redirect(r.writer, r.request, url, code)
	return nil
}

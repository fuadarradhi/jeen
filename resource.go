package jeen

import (
	"context"
	"errors"
	"net/http"
)

type Resource struct {
	// private for internal use
	writer  http.ResponseWriter
	request *http.Request

	// request
	Request *Request

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
		request: r,
		writer:  rw,

		Context: r.Context(),

		Cookie: &Cookie{
			writer:  rw,
			request: r,
		},

		Html: &Html{
			writer: rw,
			engine: h,
		},

		Json: &Json{
			writer: rw,
		},
	}
}

// Set set value to request context, see GetValue to get it
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

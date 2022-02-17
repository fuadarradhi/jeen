package jeen

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Resource struct {
	// http request wrap
	*http.Request

	// response writer
	Writer http.ResponseWriter

	// database sql.DB
	Database *Database

	// session SCS
	Session *Session

	// html response with goview
	Html *HtmlResponse

	// json response
	Json *JsonResponse
}

// create new resource
func createResource(rw http.ResponseWriter, r *http.Request, h *HtmlEngine) *Resource {
	return &Resource{
		Request: r,
		Writer:  rw,
		Html:    htmlResponse(rw, h),
		Json:    jsonResponse(rw),
	}
}

// SetValue set value to request context, see GetValue to get it
func (r *Resource) SetValue(key, val interface{}) {
	r.Request = r.Request.WithContext(context.WithValue(r.Context(), key, val))
}

// GetValue get value from context
func (r *Resource) GetValue(key interface{}) interface{} {
	return r.Context().Value(key)
}

// URLParam returns the url parameter from a http.Request object.
func (r *Resource) URLParam(key string) string {
	return chi.URLParam(r.Request, key)
}

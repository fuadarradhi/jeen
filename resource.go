package jeen

import (
	"context"
	"net/http"
)

type Resource struct {
	*http.Request
	Writer http.ResponseWriter

	Database *Database
	Session  *Session

	Html *HtmlResponse
}

func createResource(rw http.ResponseWriter, r *http.Request, t *TemplateEngine) *Resource {
	return &Resource{
		Request: r,
		Writer:  rw,
		Html:    htmlResponse(rw, t),
	}
}

func (r *Resource) SetValue(key, val interface{}) {
	r.Request = r.Request.WithContext(context.WithValue(r.Context(), key, val))
}

func (r *Resource) GetValue(key interface{}) interface{} {
	return r.Context().Value(key)
}

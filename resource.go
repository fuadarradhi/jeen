package jeen

import "net/http"

type Resource struct {
	*http.Request
	Writer http.ResponseWriter

	Database *Database
	Session  *Session
}

func createResource(rw http.ResponseWriter, r *http.Request) *Resource {
	return &Resource{
		Request: r,
		Writer:  rw,
	}
}

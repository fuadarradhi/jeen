package jeen

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	// response writer
	writer http.ResponseWriter
}

// Success is shortcut for Response with StatusOK = 200,
func (j *Json) Success(data interface{}) error {
	return j.Response(http.StatusOK, data)
}

// Error is shortcut for Response with StatusInternalServerError = 500,
func (j *Json) Error(data interface{}) error {
	return j.Response(http.StatusInternalServerError, data)
}

// Timeout is shortcut for Response with StatusGatewayTimeout = 504,
func (j *Json) Timeout(data interface{}) error {
	return j.Response(http.StatusGatewayTimeout, data)
}

// Forbidden is shortcut for Response with StatusForbidden = 403,
func (j *Json) Forbidden(data interface{}) error {
	return j.Response(http.StatusForbidden, data)
}

// NotFound is shortcut for Response with StatusNotFound = 404,
func (j *Json) NotFound(data interface{}) error {
	return j.Response(http.StatusNotFound, data)
}

// Unauthorized is shortcut for Response with StatusUnauthorized = 401,
func (j *Json) Unauthorized(data interface{}) error {
	return j.Response(http.StatusUnauthorized, data)
}

// Response response json output to browser,
func (j *Json) Response(statusCode int, data interface{}) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	j.writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	j.writer.WriteHeader(statusCode)
	_, err = j.writer.Write(out)
	return err
}

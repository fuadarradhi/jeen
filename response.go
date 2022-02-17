package jeen

import "net/http"

type HtmlResponse struct {
	// response writer
	writer http.ResponseWriter

	// template engine
	engine *TemplateEngine
}

// create new html response
func htmlResponse(rw http.ResponseWriter, t *TemplateEngine) *HtmlResponse {
	return &HtmlResponse{
		writer: rw,
		engine: t,
	}
}

// Success is shortcut for Render with StatusOK = 200,
// use escape = false if don't need html escape (default `true`)
func (h *HtmlResponse) Success(filename string, data Map, escape ...bool) error {
	return h.Render(http.StatusOK, filename, data, escape...)
}

// Error is shortcut for Render with StatusInternalServerError = 500,
// use escape = false if don't need html escape (default `true`)
func (h *HtmlResponse) Error(filename string, data Map, escape ...bool) error {
	return h.Render(http.StatusInternalServerError, filename, data, escape...)
}

// Success is shortcut for Render with StatusGatewayTimeout = 504,
// use escape = false if don't need html escape (default `true`)
func (h *HtmlResponse) Busy(filename string, data Map, escape ...bool) error {
	return h.Render(http.StatusGatewayTimeout, filename, data, escape...)
}

// Render response output to browser,
// use escape = false if don't need html escape (default `true`)
func (h *HtmlResponse) Render(statusCode int, filename string, data Map, escape ...bool) error {
	return h.engine.Render(h.writer, statusCode, filename, data, len(escape) == 0 || escape[0])
}

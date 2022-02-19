package jeen

import "net/http"

type Writer struct {
	writer http.ResponseWriter
}

func newWriter(rw http.ResponseWriter) *Writer {
	return &Writer{
		writer: rw,
	}
}

func (w *Writer) Instance() http.ResponseWriter {
	return w.writer
}

package server

import (
	"bytes"
	"net/http"
)

type hashWriter struct {
	http.ResponseWriter
	buffer     *bytes.Buffer
	hashKey    string
	statusCode int
}

func (w *hashWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *hashWriter) Write(b []byte) (int, error) {
	return w.buffer.Write(b)
}

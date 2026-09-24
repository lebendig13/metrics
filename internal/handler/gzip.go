package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w                   http.ResponseWriter
	zw                  *gzip.Writer
	responseDataWritten bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.responseDataWritten {
		c.WriteHeader(http.StatusOK)
	}
	if c.w.Header().Get("Content-Encoding") == "gzip" {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.responseDataWritten {
		return
	}
	c.responseDataWritten = true

	if statusCode < 300 {
		ct := c.w.Header().Get("Content-Type")
		if strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html") {
			c.w.Header().Set("Content-Encoding", "gzip")
		}
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	if c.w.Header().Get("Content-Encoding") == "gzip" {
		return c.zw.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

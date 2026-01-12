package ginMiddleWare

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
)

type cachedReadCloser struct {
	io.ReadCloser

	Cache bytes.Buffer

	// reader writes to Cache when it reads from io.ReaderCloser
	reader io.Reader
}

func NewCachedReadCloser(rc io.ReadCloser) *cachedReadCloser {
	c := &cachedReadCloser{ReadCloser: rc}
	c.reader = io.TeeReader(rc, &c.Cache)
	return c
}

func (rc *cachedReadCloser) Read(p []byte) (int, error) {
	return rc.reader.Read(p)
}

type cachedResponseWriter struct {
	gin.ResponseWriter

	Cache bytes.Buffer

	// writer writes to gin.ResponseWriter and Cache
	writer io.Writer
}

func NewCachedResponseWriter(w gin.ResponseWriter) *cachedResponseWriter {
	c := &cachedResponseWriter{ResponseWriter: w}
	c.writer = io.MultiWriter(w, &c.Cache)
	return c
}

func (w *cachedResponseWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

package server

import (
	"net/http"
)

func wrapLogResponseWriter(rw http.ResponseWriter) logResponseWriterIF {
	return &logResponseWriter{w: rw}
}

type logResponseWriterIF interface {
	http.ResponseWriter
	Status() int
	Size() int
}

type logResponseWriter struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *logResponseWriter) Header() http.Header {
	return l.w.Header()
}

func (l *logResponseWriter) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *logResponseWriter) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *logResponseWriter) Status() int {
	return l.status
}

func (l *logResponseWriter) Size() int {
	return l.size
}

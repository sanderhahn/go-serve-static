package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	bind := os.Getenv("BIND")
	if bind == "" {
		bind = "localhost"
	}
	addr := fmt.Sprintf("%s:%s", bind, port)

	// make files top level
	rootFs, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}

	handler := http.FileServer(http.FS(rootFs))
	handler = logger(handler)

	log.Printf("Started listening on: http://%s/\n", addr)
	err = http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatal(err)
	}
}

// Simplistic logger to display status code and duration
func logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := NewCaptureResponseWriter(w)
		inner.ServeHTTP(cw, r)
		log.Printf("%s %s %d %dms", r.Method, r.URL.Path, cw.Code, cw.Duration())
	})
}

type captureResponseWriter struct {
	http.ResponseWriter
	Code  int
	start time.Time
}

func NewCaptureResponseWriter(w http.ResponseWriter) *captureResponseWriter {
	return &captureResponseWriter{
		w,
		http.StatusOK,
		time.Now(),
	}
}

func (w *captureResponseWriter) WriteHeader(code int) {
	w.Code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureResponseWriter) Duration() time.Duration {
	return time.Since(w.start) / time.Millisecond
}

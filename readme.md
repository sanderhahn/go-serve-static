# Serve Static Embedded Files in Go

Go 1.16 supports a directive to embed files directly inside you source code:

```go
//go:embed public/*
var fs embed.FS
```

This embeds a virtual file system with an [interface](https://golang.org/pkg/embed/#pkg-index). We want to strip the `public` part of the file system so that files are directly available at top level:

```go
// make files top level
rootFs, err := fs.Sub(publicFS, "public")
```

The [http.FileServer](https://golang.org/pkg/net/http/#example_FileServer) handler can be used to expose a filesystem as http endpoint:

```go
handler := http.FileServer(http.FS(rootFs))
```

Lets add simplistic logging middleware that wraps `http.ResponseWriter` to record the status code and request duration:

```go
func logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := NewCaptureResponseWriter(w)
		inner.ServeHTTP(cw, r)
		log.Printf("%s %s %d %dms", r.Method, r.URL.Path, cw.Code, cw.Duration())
	})
}
```

The `captureResponseWriter` overrides the `WriteHeader` method to record the status code:

```go
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
```

Logging the requests using a proxy server such as [Nginx](https://docs.nginx.com/nginx/admin-guide/monitoring/logging/#setting-up-the-access-log) provides more options.
Package [httpsnoop](https://github.com/felixge/httpsnoop) claims to support more http interfaces that are used by Go handlers.

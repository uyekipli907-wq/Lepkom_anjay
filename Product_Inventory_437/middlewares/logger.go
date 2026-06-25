package middlewares

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	n, err := w.ResponseWriter.Write(data)
	w.bytes += n
	return n, err
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &loggingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}

		log.Printf(
			"%s %s %d %s %dB",
			r.Method,
			r.URL.Path,
			recorder.status,
			time.Since(start).Round(time.Millisecond),
			recorder.bytes,
		)
	})
}

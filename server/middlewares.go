package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// Key to use when setting the request id.
type ctxKeyRequestID int

// requestIDKey is the key that holds the unique request id in a request context.
const requestIDKey ctxKeyRequestID = iota

// responseRecorder records response status code that wraps http.ResponseWriter.
type responseRecorder struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.code = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// loggingMiddleware is a middleware that logs the start and end of each request, along with other
// useful data like request_id from request and user_id from token if available.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &responseRecorder{ResponseWriter: w}
		start := time.Now()

		defer func() {
			ctx := r.Context()
			m := fmt.Sprintf("%d %s %s %s", ww.code, r.Method, r.URL, time.Since(start))
			reqID, _ := ctx.Value(requestIDKey).(string)
			s.logger.InfoContext(ctx, m,
				"path", r.URL.EscapedPath(),
				"method", r.Method,
				"status", slog.IntValue(ww.code),
				"duration_ms", slog.Int64Value(time.Since(start).Milliseconds()),
				"request_id", reqID,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

// recoveryMiddleware is a middleware that handles panic and recovers them, it also
// logs request dump and stack trace.
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				rqd, err := httputil.DumpRequest(r, true)
				if err != nil {
					s.logger.Error(err.Error())
					encodeJSONError(w, err, http.StatusInternalServerError)
					return
				}

				s.logger.Error(fmt.Sprintf("panic: %+v\n\nrequest dump: %s\nstack trace: %s",
					rvr, rqd, debug.Stack()))
				fmt.Printf("%s", debug.Stack())
				encodeJSONError(w,
					fmt.Errorf(http.StatusText(http.StatusInternalServerError)),
					http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

const requestIDHeaderKey = "X-Request-Id"

// requestIDMiddleware is a middleware that sets X-Request-Id from header to request context and
// response header. request id will be generated if not available from request.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(requestIDHeaderKey)
		if rid == "" {
			rid = uuid.NewString()
		}

		w.Header().Set(requestIDHeaderKey, rid)
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

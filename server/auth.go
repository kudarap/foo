package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// authentication is middleware that looks for authorization bearer token
// from header request and process verification. Verified token provides authorized
// user id and adds to request context that can be use for validating authorized requests.
//
// When authorization header is not present it skips the verification.
func authentication(auth authenticator) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token, err := availableTokenFromHeader(r.Header)
			if err != nil {
				encodeJSONError(w, err, http.StatusForbidden)
				return
			}
			// When token is available it will be validated and parsed to get user id
			// and attach to context for next handler.
			if token != "" {
				ctx := r.Context()
				claims, err := auth.VerifyToken(ctx, token)
				if err != nil {
					encodeJSONError(w, err, http.StatusForbidden)
					return
				}
				userID, _ := claims["user_id"].(string)
				ctx = userToContext(ctx, userID)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// authorizedMiddleware is a middleware that requires authorized user id from a request.
// An empty user ID will result to unauthorized request 403.
func authorizedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if userFromContext(ctx) == "" {
			encodeJSONError(w, fmt.Errorf("unauthorized request"), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Key to use when setting the user id.
type ctxKeyUserID int

// requestIDKey is the key that holds the unique user id in a request context.
const userIDKey ctxKeyUserID = iota

// userToContext sets user id to context.
func userToContext(parent context.Context, userID string) context.Context {
	return context.WithValue(parent, userIDKey, userID)
}

// userFromContext returns user id from the context if one is present.
func userFromContext(ctx context.Context) (userID string) {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

type authenticator interface {
	VerifyToken(ctx context.Context, token string) (claims map[string]interface{}, err error)
}

// availableTokenFromHeader checks Authorization value from header and extracts token. Returns empty string
// and nil error when Authorization has no value.
func availableTokenFromHeader(h http.Header) (token string, err error) {
	ah := h.Get("Authorization")
	if ah == "" {
		return "", nil
	}

	t := strings.TrimPrefix(ah, "Bearer ")
	if t == "" {
		return "", errors.New("malformed authInjector header")
	}
	return t, nil
}

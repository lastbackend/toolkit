package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

// KeyPEMBlock to use when setting the request UserID.
type ctxKeyRequestID int

// RequestIDKey is the key that holds th unique request UserID in a request context.
const RequestIDKey ctxKeyRequestID = 0

var (
	// prefix is const prefix for request UserID
	prefix string

	// reqID is counter for request UserID
	reqID uint64
)

// init Initializes constant part of request UserID
func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		_, _ = rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

// RequestID is a middleware that injects a request UserID into the context of each
// request. A request UserID is a string of the form "host.example.com/random-0001",
// where "random" is a base62 random string that uniquely identifies this go
// process, and where the last number is an atomically incremented request
// counter.
func (m Middleware) RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := atomic.AddUint64(&reqID, 1)
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDKey, fmt.Sprintf("%s-%06d", prefix, id))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetReqID returns a request UserID from the given context if one is present.
// Returns the empty string if a request UserID cannot be found.
func (m Middleware) GetReqID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

package proxy

import (
	"context"
	"net/http"
	"sync"

	l "github.com/empijei/web-proxy/log"
	ulid "github.com/oklog/ulid/v2"
)

// Action represents the possible actions interceptors can take.
type Action int64

const (
	// ActionContinue makes the proxy continue executing the request flow, going through
	// the following interceptors.
	ActionContinue Action = iota
	// ActionDrop makes the proxy stop and generate a Bad Gateway response.
	ActionDrop
	// ActionSkip makes the proxy skip all the remaining interceptors and just execute
	// the request/response flight.
	ActionSkip
)

type (
	// RequestInterceptor is a function that can modify and potentially drop a request.
	RequestInterceptor func(ctx context.Context, rt *RoundTrip, req *http.Request) Action
	// ResponseInterceptor is a function that can modify and potentially drop a response.
	ResponseInterceptor func(ctx context.Context, rt *RoundTrip, resp *http.Response) Action
)

// RoundTrip is contextual data related to a request-response roundtrip.
type RoundTrip struct {
	// ID is the identifier for the roundtrip.
	ID    ulid.ULID
	store sync.Map
}

// RoundTripKey is a typed key to store and load values from a roundtrip.
type RoundTripKey[T any] string

// Set sets the value for the key.
func (rtk RoundTripKey[T]) Set(rt *RoundTrip, value T) {
	rt.store.Store(rtk, value)
}

// Get retrieves the value for the key.
func (rtk RoundTripKey[T]) Get(rt *RoundTrip) (value T, ok bool) {
	v, ok := rt.store.Load(rtk)
	if !ok {
		return value, ok
	}
	value, ok = v.(T)
	if !ok {
		l.Fatalf("value for key %q should be %T but was %T", rtk, value, v)
	}
	return value, ok
}

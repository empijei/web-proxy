package proxy

import (
	"net/http"

	l "github.com/empijei/web-proxy/log"
)

type (
	// RequestInterceptorMiddleWare is a type use to wrap an inner request interceptor.
	// This allows decoration logic to perform operations like logging or deciding whether
	// to run the next interceptor in the chain.
	RequestInterceptorMiddleWare func(RequestInterceptor) RequestInterceptor
	// RequestInterceptor is a function that can modify a request.
	//
	// If a response is returned, the request is not sent.
	RequestInterceptor func(rt *RoundTrip, req *http.Request) *http.Response

	// ResponseInterceptorMiddleWare is a type use to wrap an inner response interceptor.
	// This allows decoration logic to perform operations like logging or deciding whether
	// to run the next interceptor in the chain.
	ResponseInterceptorMiddleWare func(ResponseInterceptor) ResponseInterceptor
	// ResponseInterceptor is a function that can modify a response.
	//
	// Response interceptors are called on generated responses, with rt.Skipped
	// set to true.
	ResponseInterceptor func(rt *RoundTrip, resp *http.Response)
)

type RoundTripID uint64

// RoundTrip is contextual data related to a request-response roundtrip.
type RoundTrip struct {
	// Fields set by proxy:

	// ProxyName is the name of the proxy that intercepted this roundtrip.
	ProxyName string
	// ID is the identifier for the roundtrip.
	ID RoundTripID
	// Skipped is set to true by the proxy if the request never hit the server,
	// but a response was generated instead.
	Skipped bool

	// Fields set by interceptors:

	// RequestEdited MUST be set when an interceptor modifies a request.
	RequestEdited bool
	// ResponseEdited MUST be set when an interceptor modifies a response.
	ResponseEdited bool

	store map[any]any
}

// NewRoundTrip can be used to construct a RoundTrip for testing or for use outside
// of the proxy.
func (p *Proxy) NewRoundTrip() *RoundTrip {
	return &RoundTrip{
		ProxyName: p.name,
		ID:        RoundTripID(p.ids.Add(1)),
	}
}

// RoundTripKey is a typed key to store and load values from a roundtrip.
//
// Ideally K should be an unexported type to make sure there is no collision
// between packages trying to set the same key.
//
// Example:
//
//	type myKey struct{}
//	var MyRoundTripKey = RoundTripKey[myKey, string]{}
type RoundTripKey[K comparable, V any] struct{}

// Set sets the value for the key.
func (rtk RoundTripKey[K, T]) Set(rt *RoundTrip, value T) {
	if rt.store == nil {
		rt.store = map[any]any{}
	}
	var k K
	rt.store[k] = value
}

// Get retrieves the value for the key.
func (rtk RoundTripKey[K, T]) Get(rt *RoundTrip) (value T, ok bool) {
	var k K
	v, ok := rt.store[k]
	if !ok {
		return value, ok
	}
	value, ok = v.(T)
	if !ok {
		l.Fatalf("value for key %q should be %T but was %T", rtk, value, v)
	}
	return value, ok
}

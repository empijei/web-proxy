package proxy

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/elazarl/goproxy"
	l "github.com/empijei/web-proxy/log"
	ulid "github.com/oklog/ulid/v2"
)

type (
	// RequestInterceptor is a function that can modify and potentially drop a request.
	RequestInterceptor func(ctx context.Context, rt *RoundTrip, req *http.Request) (keep bool)
	// ResponseInterceptor is a function that can modify and potentially drop a response.
	ResponseInterceptor func(ctx context.Context, rt *RoundTrip, resp *http.Response) (keep bool)
)

// Proxy is a machine-in-the-middle proxy.
type Proxy struct {
	name string
	gp   *goproxy.ProxyHttpServer
	addr string

	started  *atomic.Bool
	reqMitm  []RequestInterceptor
	respMitm []ResponseInterceptor
}

// New returns a new proxy using ca as Certificate Authority.
func New(ca *tls.Certificate, addr, name string) (*Proxy, error) {
	p := &Proxy{
		name: name,
		gp:   goproxy.NewProxyHttpServer(),
		addr: addr,
	}

	customCaMitm := &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(ca)}
	p.gp.OnRequest().HandleConnect(
		goproxy.FuncHttpsHandler(func(host string, _ *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			return customCaMitm, host
		}))

	p.gp.OnRequest().DoFunc(p.onReq)
	p.gp.OnResponse().DoFunc(p.onResp)
	return p, nil
}

func (p *Proxy) onReq(req *http.Request, gctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	rt := &RoundTrip{ID: ulid.Make()}
	gctx.UserData = rt
	ctx := req.Context()
	for _, f := range p.reqMitm {
		if !f(ctx, rt, req) {
			return nil, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "Request dropped")
		}
	}
	req = req.WithContext(ctx)
	return req, nil
}

// Intercept allows to setup additional request or response interception logic.
//
// Intercept MUST be called before Serve.
func (p *Proxy) Intercept(req RequestInterceptor, resp ResponseInterceptor) {
	if p.started.Load() {
		panic("cannot call Intercept after the proxy.Handler has been created")
	}
	if req != nil {
		p.reqMitm = append(p.reqMitm, req)
	}
	if resp != nil {
		p.respMitm = append(p.respMitm, resp)
	}
}

// Serve starts the proxy.
func (p *Proxy) Serve(ctx context.Context) {
	p.started.Store(true)
	s := &http.Server{ //nolint: gosec // this is running locally.
		Addr:    p.addr,
		Handler: p.gp,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			l.Errorf("Proxy %q Serve: %v", p.name, err)
		}
	}()
	context.AfterFunc(ctx, func() {
		_ = s.Close()
	})
}

func (p *Proxy) onResp(resp *http.Response, gctx *goproxy.ProxyCtx) *http.Response {
	rt, ok := gctx.UserData.(*RoundTrip)
	ctx := resp.Request.Context()
	if !ok {
		l.Errorf("Got response without RoundTrip for '%s %s'", resp.Request.Method, resp.Request.URL.Path)
		return resp
	}
	for _, f := range p.respMitm {
		if !f(ctx, rt, resp) {
			return goproxy.NewResponse(resp.Request, goproxy.ContentTypeText, http.StatusBadGateway, "Response dropped")
		}
	}
	resp.Request = resp.Request.WithContext(ctx)
	return resp
}

/*

func ParseCA(caCert, caKey []byte) (*tls.Certificate, error) {
	parsedCert, err := tls.X509KeyPair(caCert, caKey)
	if err != nil {
		return nil, err
	}
	if parsedCert.Leaf, err = x509.ParseCertificate(parsedCert.Certificate[0]); err != nil {
		return nil, err
	}
	return &parsedCert, nil
}
*/

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

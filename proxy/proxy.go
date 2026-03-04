// Package proxy implements a mitm https proxy with support for pluggable intercept logic.
package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"sync/atomic"

	"github.com/elazarl/goproxy"
	l "github.com/empijei/web-proxy/log"
	ulid "github.com/oklog/ulid/v2"
)

// Proxy is a machine-in-the-middle proxy.
type Proxy struct {
	name string
	gp   *goproxy.ProxyHttpServer

	started  atomic.Bool
	reqMitm  []RequestInterceptor
	respMitm []ResponseInterceptor
}

// New returns a new proxy using ca as Certificate Authority.
func New(ca *tls.Certificate, name string) (*Proxy, error) {
	p := &Proxy{
		name: name,
		gp:   goproxy.NewProxyHttpServer(),
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

// Intercept allows to setup additional request or response interception logic.
//
// Intercept MUST be called before starting the proxy.
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

// Handler returns the handler running the CONNECT proxy.
func (p *Proxy) Handler() http.Handler {
	p.started.Store(true)
	return p.gp
}

var skipKey = RoundTripKey[bool]("proxy:skip")

func (p *Proxy) onReq(req *http.Request, gctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	rt := &RoundTrip{ID: ulid.Make()}
	gctx.UserData = rt
	ctx := req.Context()
	for _, f := range p.reqMitm {
		switch f(ctx, rt, req) {
		case ActionSkip:
			skipKey.Set(rt, true)
			return req, nil
		case ActionDrop:
			return nil, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusBadGateway, "Request dropped")
		case ActionContinue:
		}
	}
	req = req.WithContext(ctx)
	return req, nil
}

func (p *Proxy) onResp(resp *http.Response, gctx *goproxy.ProxyCtx) *http.Response {
	rt, ok := gctx.UserData.(*RoundTrip)
	if !ok {
		l.Errorf("Got response without RoundTrip for '%s %s'", resp.Request.Method, resp.Request.URL.Path)
		return resp
	}
	if skip, ok := skipKey.Get(rt); ok && skip {
		return resp
	}

	ctx := resp.Request.Context()
	for _, f := range p.respMitm {
		switch f(ctx, rt, resp) {
		case ActionSkip:
			return resp
		case ActionDrop:
			return goproxy.NewResponse(resp.Request, goproxy.ContentTypeText, http.StatusBadGateway, "Response dropped")
		case ActionContinue:
		}
	}
	resp.Request = resp.Request.WithContext(ctx)
	return resp
}

// ParseCA parses the given cert and key.
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

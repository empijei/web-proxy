// Package proxy implements a mitm https proxy with support for pluggable intercept logic.
package proxy

import (
	"crypto/tls"
	"net/http"
	"sync/atomic"

	"github.com/elazarl/goproxy"
	l "github.com/empijei/web-proxy/log"
)

// Proxy is a machine-in-the-middle proxy.
type Proxy struct {
	name string
	gp   *goproxy.ProxyHttpServer

	started  atomic.Bool
	reqMitm  RequestInterceptor
	respMitm ResponseInterceptor
}

// New returns a new proxy using ca as Certificate Authority.
func New(ca *tls.Certificate, name string) (*Proxy, error) {
	p := &Proxy{
		name: name,
		gp:   goproxy.NewProxyHttpServer(),

		// The base intercptors do nothing.
		reqMitm:  func(_ *RoundTrip, _ *http.Request) *http.Response { return nil },
		respMitm: func(_ *RoundTrip, _ *http.Response) {},
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
func (p *Proxy) Intercept(req RequestInterceptorMiddleWare, resp ResponseInterceptorMiddleWare) {
	if p.started.Load() {
		panic("cannot call Intercept after the proxy.Handler has been created")
	}
	if req != nil {
		p.reqMitm = req(p.reqMitm)
	}
	if resp != nil {
		p.respMitm = resp(p.respMitm)
	}
}

// Handler returns the handler running the CONNECT proxy.
func (p *Proxy) Handler() http.Handler {
	p.started.Store(true)
	return p.gp
}

func (p *Proxy) onReq(req *http.Request, gctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	rt := NewRoundTrip(p.name)
	gctx.UserData = rt
	resp := p.reqMitm(rt, req)
	if resp != nil {
		rt.Skipped = true
		p.onResp(resp, gctx)
		return nil, resp
	}
	return req, nil
}

func (p *Proxy) onResp(resp *http.Response, gctx *goproxy.ProxyCtx) *http.Response {
	rt, ok := gctx.UserData.(*RoundTrip)
	if !ok {
		l.Errorf("Got response without RoundTrip for '%s %s'", resp.Request.Method, resp.Request.URL.Path)
		return resp
	}
	p.respMitm(rt, resp)
	return resp
}

package proxy_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/proxy"
	"github.com/empijei/web-proxy/testing/proxytesting"
	"github.com/empijei/web-proxy/ui"
)

func TestProxy(t *testing.T) {
	tst.Go(t)
	cert, ca := proxytesting.SetupCert(t)
	p := tst.Do(proxy.New(cert, "test"))(t)
	var (
		gotReq      *http.Request
		gotRespBody string
		gotReqID    ui.RoundTripID
		gotRespID   ui.RoundTripID
		gotVal      int
		gotProxy    string
	)
	type pkey struct{}
	rk := proxy.RoundTripKey[pkey, int]{}
	p.Intercept(func(ri proxy.RequestInterceptor) proxy.RequestInterceptor {
		return func(rt *proxy.RoundTrip, req *http.Request) *http.Response {
			gotProxy = rt.ProxyName
			gotReq = req
			gotReqID = rt.ID
			rk.Set(rt, 42)
			return ri(rt, req)
		}
	}, func(ri proxy.ResponseInterceptor) proxy.ResponseInterceptor {
		return func(rt *proxy.RoundTrip, resp *http.Response) {
			buf, _ := io.ReadAll(resp.Body)
			gotRespBody = string(buf)
			resp.Body = io.NopCloser(strings.NewReader(gotRespBody))

			gotRespID = rt.ID
			gotVal, _ = rk.Get(rt)
			ri(rt, resp)
		}
	})

	msg := `Hello, World!`
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, msg)
	})
	remote, cl := proxytesting.SetupProxyAndClient(t, ca, p, h)

	tst.Is("https", tst.Do(url.Parse(
		remote))(t).Scheme, t)
	resp := tst.Do(cl.Get(remote))(t)
	tst.Is(http.StatusOK, resp.StatusCode, t)
	tst.Is(msg, string(tst.Do(io.ReadAll(
		resp.Body))(t)), t)
	tst.Is(http.MethodGet, gotReq.Method, t)
	tst.Is(msg, gotRespBody, t)
	tst.Is(42, gotVal, t)
	tst.Is(gotReqID, gotRespID, t)
	tst.Is("test", gotProxy, t)
}

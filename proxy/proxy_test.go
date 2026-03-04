package proxy_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/proxy"
	ulid "github.com/oklog/ulid/v2"
)

var (
	//go:embed testdata/cert.pem
	caCert []byte
	//go:embed testdata/key.pem
	caKey []byte
)

func setupCert(t *testing.T) (ca *tls.Certificate, caCertPool *x509.CertPool) {
	t.Helper()
	rootCert := tst.Do(proxy.ParseCA(caCert, caKey))(t)

	caCertPool = x509.NewCertPool()
	leaf := tst.Do(x509.ParseCertificate(rootCert.Certificate[0]))(t)
	caCertPool.AddCert(leaf)
	return rootCert, caCertPool
}

func setupClient(t *testing.T, caCertPool *x509.CertPool, proxyURL string) *http.Client {
	t.Helper()
	u := tst.Do(url.Parse(proxyURL))(t)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
		Proxy: func(_ *http.Request) (*url.URL, error) {
			return u, nil
		},
	}
	return &http.Client{
		Transport: transport,
	}
}

func TestProxy(t *testing.T) {
	tst.Go(t)
	cert, ca := setupCert(t)
	p := tst.Do(proxy.New(cert, "test"))(t)
	var (
		gotReq      *http.Request
		gotRespBody string
		gotReqID    ulid.ULID
		gotRespID   ulid.ULID
		gotVal      int
	)
	rk := proxy.RoundTripKey[int]("proxy_test:base")
	p.Intercept(func(ctx context.Context, rt *proxy.RoundTrip, req *http.Request) proxy.Action {
		gotReq = req

		gotReqID = rt.ID
		rk.Set(rt, 42)
		return proxy.ActionContinue
	}, func(ctx context.Context, rt *proxy.RoundTrip, resp *http.Response) proxy.Action {
		buf, _ := io.ReadAll(resp.Body)
		gotRespBody = string(buf)
		resp.Body = io.NopCloser(strings.NewReader(gotRespBody))

		gotRespID = rt.ID
		gotVal, _ = rk.Get(rt)
		return proxy.ActionContinue
	})

	mitm := httptest.NewServer(p.Handler())
	defer mitm.Close()

	cl := setupClient(t, ca, mitm.URL)

	msg := `Hello, World!`
	endpoint := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, msg)
	}))
	endpoint.StartTLS()
	defer endpoint.Close()

	tst.Is("https", tst.Do(url.Parse(
		endpoint.URL))(t).Scheme, t)
	resp := tst.Do(cl.Get(endpoint.URL))(t)
	tst.Is(http.StatusOK, resp.StatusCode, t)
	tst.Is(msg, string(tst.Do(io.ReadAll(
		resp.Body))(t)), t)
	tst.Is(http.MethodGet, gotReq.Method, t)
	tst.Is(msg, gotRespBody, t)
	tst.Is(42, gotVal, t)
	tst.Is(gotReqID, gotRespID, t)
}

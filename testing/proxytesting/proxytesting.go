package proxytesting

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	_ "embed"

	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/proxy"
)

var (
	//go:embed testdata/cert.pem
	caCert []byte
	//go:embed testdata/key.pem
	caKey []byte
)

// SetupCert sets up a simple certificate authority.
func SetupCert(t tst.Test) (ca *tls.Certificate, caCertPool *x509.CertPool) {
	t.Helper()
	rootCert := tst.Do(proxy.ParseCA(caCert, caKey))(t)

	caCertPool = x509.NewCertPool()
	leaf := tst.Do(x509.ParseCertificate(rootCert.Certificate[0]))(t)
	caCertPool.AddCert(leaf)
	return rootCert, caCertPool
}

// SetupProxyAndClient starts the proxy and creates a client that trusts its certs an uses it as a proxy.
func SetupProxyAndClient(t tst.Test, caCertPool *x509.CertPool, p *proxy.Proxy, remoteHandler http.Handler) (remoteURL string, cl *http.Client) {
	t.Helper()
	mitm := httptest.NewServer(p.Handler())
	t.Cleanup(func() {
		mitm.Close()
	})
	u := tst.Do(url.Parse(mitm.URL))(t)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
		Proxy: func(_ *http.Request) (*url.URL, error) {
			return u, nil
		},
	}

	if remoteHandler == nil {
		remoteHandler = EchoHandler
	}
	endpoint := httptest.NewUnstartedServer(remoteHandler)
	endpoint.StartTLS()
	t.Cleanup(func() {
		endpoint.Close()
	})
	return endpoint.URL, &http.Client{
		Transport: transport,
	}
}

// EchoHandler is a http.Handler that echoes the request body.
var EchoHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(w, r.Body)
})

package proxy_test

import (
	"testing"

	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/proxy"
)

func TestCA(t *testing.T) {
	tst.Go(t)
	dir := t.TempDir()
	caCertTmp, caKeyTmp := tst.Do2(
		proxy.GenerateCA())(t)
	tst.No(
		proxy.StoreCA(dir, caCertTmp, caKeyTmp), t)

	caCert, caKey := tst.Do2(
		proxy.LoadCA(dir))(t)
	_ = tst.Do(
		proxy.ParseCA(caCert, caKey))(t)

	tst.Is(string(caCertTmp), string(caCert), t)
	tst.Is(string(caKeyTmp), string(caKey), t)
}

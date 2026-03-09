// Package wapty implements a simple Web Application Penetration Testing suite.
package wapty

import (
	"context"
	"fmt"
	"net/http"

	"github.com/empijei/web-proxy/history"
	l "github.com/empijei/web-proxy/log"
	"github.com/empijei/web-proxy/proxy"
)

// Run runs a Wep Application Penetration Testing suite.
func Run(ctx context.Context, certDir string) error {
	cert, key, err := proxy.LoadCA(certDir)
	if err != nil {
		cert, key, err = proxy.GenerateCA()
		if err != nil {
			return fmt.Errorf("generate CA: %w", err)
		}
		if err := proxy.StoreCA(certDir, cert, key); err != nil {
			return fmt.Errorf("store CA: %w", err)
		}
		l.Infof("Certificates generated")
	}
	ca, err := proxy.ParseCA(cert, key)
	if err != nil {
		return fmt.Errorf("parse CA: %w", err)
	}

	l.Infof("Certificates available in %s", certDir)

	p, err := proxy.New(ca, "wapty:default")
	if err != nil {
		return fmt.Errorf("create proxy: %w", err)
	}
	r := history.NewRecorder()
	p.Intercept(r.MiddleWare())
	evts := r.Events()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-evts:
				fmt.Printf("%d: %s %s://%s%s -> %v %v", e.Metadata.ID,
					e.Metadata.Method, e.Metadata.Scheme, e.Metadata.Host, e.Metadata.PathAndQuery, e.Metadata.StatusCode, e.Metadata.ContentType)
			}
		}
	}()
	addr := "localhost:8989"
	l.Infof("Proxy listening on %q", addr)

	return fmt.Errorf("proxy listen: %w", http.ListenAndServe(addr, p.Handler())) //nolint: gosec // This is temporary
}

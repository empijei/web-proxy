// Package api implements the UI API.
package api

import (
	"context"

	"github.com/empijei/srpc"
	"github.com/empijei/web-proxy/history"
)

// Setup registers the API on the given mux.
func Setup(ctx context.Context, mux srpc.Mux, r *history.Recorder) {
	r.RegisterAPI(ctx, mux)
}

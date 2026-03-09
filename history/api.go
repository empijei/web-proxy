package history

import (
	"context"
	"iter"

	"github.com/empijei/chans"
	"github.com/empijei/srpc"
	l "github.com/empijei/web-proxy/log"
	"github.com/empijei/web-proxy/ui"
)

// RegisterAPI registers the history endpoints on the API.
//
// MUST be called only once per Recorder.
func (r *Recorder) RegisterAPI(ctx context.Context, mux srpc.Mux) {
	evts := r.Events()
	buf := chans.Unbound(ctx.Done(), evts, 1024, func(aboveThreshold bool) {
		if aboveThreshold {
			l.Warnf("UI can't keep up with history metadata!")
			return
		}
		l.Infof("UI resumed keeping up")
	})
	multi := chans.NewMulticast(ctx.Done(), buf)

	ui.HistoryMetadataEP.Register(mux, func(ctx context.Context, _ ui.HistoryMetadataRequest) (iter.Seq2[ui.HistoryMetadataResponse, error], error) {
		r.mu.Lock()
		sub := multi.Subscribe(ctx.Done(), 1)
		prev := r.unsafeGetUntilLast(ctx)
		r.mu.Unlock()
		all := chans.Concat(ctx.Done(), prev, sub)
		return func(yield func(ui.HistoryMetadataResponse, error) bool) {
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-all:
					if !ok {
						return
					}
					if !yield(v.Metadata, nil) {
						return
					}
				}
			}
		}, nil
	})
}

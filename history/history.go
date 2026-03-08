package history

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"path"
	"sync"
	"sync/atomic"
	"time"

	l "github.com/empijei/web-proxy/log"
	"github.com/empijei/web-proxy/proxy"
	"github.com/empijei/web-proxy/ui"
	"github.com/oklog/ulid/v2"
)

// Entry is an entry in the history.
type Entry struct {
	Metadata         ui.TrafficOverview
	originalRequest  string
	editedRequest    string
	originalResponse string
	editedResponse   string
}

func (e *Entry) OriginalRequest() string  { return e.originalRequest }
func (e *Entry) EditedRequest() string    { return e.editedRequest }
func (e *Entry) OriginalResponse() string { return e.originalResponse }
func (e *Entry) EditedResponse() string   { return e.editedResponse }

func overViewString(to ui.TrafficOverview) string {
	return fmt.Sprintf("%s %s://%s", to.Method, to.Scheme, path.Join(to.Host, to.PathAndQuery))
}

// Recorder allows to record proxy history.
type Recorder struct {
	now   func() time.Time
	close chan struct{}

	mu    sync.RWMutex
	state map[ulid.ULID]*Entry

	generateEvts atomic.Bool
	evt          chan Entry
}

// NewRecorder constructs a new recorder.
func NewRecorder() *Recorder {
	return &Recorder{
		now:   time.Now,
		close: make(chan struct{}),
		state: map[ulid.ULID]*Entry{},
	}
}

// MiddleWare returns the middleware to use on a proxy Intercept.
func (r *Recorder) MiddleWare() (proxy.RequestInterceptorMiddleWare, proxy.ResponseInterceptorMiddleWare) {
	return r.onReq, r.onResp
}

// Get returns the specified entry.
func (r *Recorder) Get(roundtripID ulid.ULID) (_ Entry, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.state[roundtripID]
	if !ok {
		return Entry{}, false
	}
	return *e, ok
}

// GetAll returns the entire state.
func (r *Recorder) GetAll() []Entry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ret []Entry
	for _, e := range r.state {
		ret = append(ret, *e)
	}
	return ret
}

// Events return the events channel. Only one consumer should be reading events
// from the returned channel, multi-casting should be done by the caller.
//
// If consumers are too slow at processing events, the recorder will block.
func (r *Recorder) Events() <-chan Entry {
	if !r.generateEvts.CompareAndSwap(false, true) {
		panic("(*history.Recoder).Events() called multiple times")
	}
	r.evt = make(chan Entry)
	return r.evt
}

// Stop stops the recorder.
func (r *Recorder) Stop() {
	close(r.close)
}

func (r *Recorder) onReq(ri proxy.RequestInterceptor) proxy.RequestInterceptor {
	return func(rt *proxy.RoundTrip, req *http.Request) *http.Response {
		var qs string
		if req.URL.RawQuery != "" {
			qs = "?"
		}
		e := &Entry{
			Metadata: ui.TrafficOverview{
				ID:           rt.ID,
				Scheme:       req.URL.Scheme,
				Host:         req.Host,
				Method:       req.Method,
				PathAndQuery: req.URL.Path + qs + req.URL.RawQuery,
				StartedAt:    r.now(),
				ProxyName:    rt.ProxyName,
			},
		}
		buf, err := httputil.DumpRequest(req, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump request: %q: %v", overViewString(e.Metadata), err)
		}
		e.originalRequest = string(buf)

		{
			r.mu.Lock()
			r.state[rt.ID] = e
			r.mu.Unlock()
		}

		resp := ri(rt, req)

		if !rt.RequestEdited {
			return resp
		}

		buf, err = httputil.DumpRequest(req, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump modified request: %q: %v", overViewString(e.Metadata), err)
		}
		{
			r.mu.Lock()
			e.Metadata.RequestEdited = true
			e.editedRequest = string(buf)
			r.mu.Unlock()
		}
		return resp
	}
}

func (r *Recorder) onResp(ri proxy.ResponseInterceptor) proxy.ResponseInterceptor {
	return func(rt *proxy.RoundTrip, resp *http.Response) {
		buf, err := httputil.DumpResponse(resp, true)

		r.mu.Lock()
		e := r.state[rt.ID]
		e.Metadata.StatusCode = resp.StatusCode
		e.Metadata.ContentType = resp.Header.Get("Content-Type")
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump response: %q: %v", overViewString(e.Metadata), err)
		}
		e.originalResponse = string(buf)
		r.mu.Unlock()

		if r.generateEvts.Load() {
			defer func() {
				select {
				case r.evt <- *e:
				case <-r.close:
				}
			}()
		}

		ri(rt, resp)

		if !rt.ResponseEdited {
			return
		}
		buf, err = httputil.DumpResponse(resp, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump modified response: %q: %v", overViewString(e.Metadata), err)
		}

		r.mu.Lock()
		e.Metadata.ResponseEdited = true
		e.editedResponse = string(buf)
		r.mu.Unlock()
	}
}

package history

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"path"
	"sync"
	"time"

	l "github.com/empijei/web-proxy/log"
	"github.com/empijei/web-proxy/proxy"
	"github.com/empijei/web-proxy/ui"
	"github.com/oklog/ulid/v2"
)

type Entry struct {
	Metadata         ui.TrafficOverview
	OriginalRequest  []byte
	OriginalResponse []byte
	EditedRequest    []byte
	EditedResponse   []byte
}

func overViewString(to ui.TrafficOverview) string {
	return fmt.Sprintf("%s %s://%s", to.Method, to.Scheme, path.Join(to.Host, to.PathAndQuery))
}

type Recorder struct {
	mu    sync.RWMutex
	state map[ulid.ULID]*Entry
	now   func() time.Time
}

func NewRecorder() *Recorder {
	return &Recorder{
		state: map[ulid.ULID]*Entry{},
		now:   time.Now,
	}
}

func (r *Recorder) MiddleWare() (proxy.RequestInterceptorMiddleWare, proxy.ResponseInterceptorMiddleWare) {
	return r.onReq, r.onResp
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
		buf, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump request: %q: %v", overViewString(e.Metadata), err)
		}
		e.OriginalRequest = buf

		r.mu.Lock()
		r.state[rt.ID] = e
		r.mu.Unlock()

		resp := ri(rt, req)

		if !rt.RequestEdited {
			return resp
		}

		e.Metadata.RequestEdited = true
		buf, err = httputil.DumpRequestOut(req, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump modified request: %q: %v", overViewString(e.Metadata), err)
		}
		e.EditedRequest = buf
		return resp
	}
}

func (r *Recorder) onResp(ri proxy.ResponseInterceptor) proxy.ResponseInterceptor {
	return func(rt *proxy.RoundTrip, resp *http.Response) {
		r.mu.Lock()
		e := r.state[rt.ID]
		r.mu.Unlock()

		e.Metadata.StatusCode = resp.StatusCode
		e.Metadata.ContentType = resp.Header.Get("Content-Type")
		buf, err := httputil.DumpResponse(resp, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump response: %q: %v", overViewString(e.Metadata), err)
		}
		e.OriginalResponse = buf

		ri(rt, resp)

		if !rt.ResponseEdited {
			return
		}
		e.Metadata.ResponseEdited = true
		buf, err = httputil.DumpResponse(resp, true)
		if err != nil {
			buf = nil
			l.Errorf("Cannot dump modified response: %q: %v", overViewString(e.Metadata), err)
		}
		e.EditedResponse = buf
	}
}

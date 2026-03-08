package history_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/history"
	"github.com/empijei/web-proxy/proxy"
	"github.com/empijei/web-proxy/testing/proxytesting"
	"github.com/empijei/web-proxy/ui"
)

type stubInnerInterceptor struct {
	id         proxy.RoundTripID
	modifyReq  bool
	modifyResp bool
}

func (s *stubInnerInterceptor) mw() (proxy.RequestInterceptorMiddleWare, proxy.ResponseInterceptorMiddleWare) {
	return func(ri proxy.RequestInterceptor) proxy.RequestInterceptor {
			return func(rt *proxy.RoundTrip, req *http.Request) *http.Response {
				s.id = rt.ID
				if s.modifyReq {
					rt.RequestEdited = true
				}
				req.Header.Set("X-Request-Modified", "true")
				return ri(rt, req)
			}
		}, func(ri proxy.ResponseInterceptor) proxy.ResponseInterceptor {
			return func(rt *proxy.RoundTrip, resp *http.Response) {
				if s.modifyResp {
					rt.ResponseEdited = true
				}
				resp.Header.Set("X-Response-Modified", "true")
				ri(rt, resp)
			}
		}
}

func TestMiddleWareSingleFlight(t *testing.T) {
	tst.Go(t)
	now := time.Now()
	r := history.NewRecorder()
	r.SetClock(func() time.Time {
		return now
	})

	var evt history.Entry
	{
		evts := r.Events()
		defer r.Stop()
		go func() {
			for {
				select {
				case evt = <-evts:
				case <-t.Context().Done():
				}
			}
		}()
	}

	ca, caPool := proxytesting.SetupCert(t)
	p := tst.Do(proxy.New(ca, "test:history"))(t)
	sii := &stubInnerInterceptor{modifyReq: true, modifyResp: true}
	p.Intercept(sii.mw())
	p.Intercept(r.MiddleWare())

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, "bar")
	})
	remote, cl := proxytesting.SetupProxyAndClient(t, caPool, p, h)

	hresp := tst.Do(cl.Post(remote+"/foo?q=42", "text/plain", strings.NewReader("foo")))(t)
	ru := tst.Do(url.Parse(remote))(t)

	got := r.GetAll()
	tst.Be(len(got) == 1, t)
	e := got[0]
	t.Logf("entry:\n%s", tst.Do(json.MarshalIndent(e, "", "\t"))(t))
	tst.Is(ui.TrafficOverview{
		ID:             uint64(sii.id),
		Scheme:         "https",
		Host:           ru.Host,
		Method:         "POST",
		PathAndQuery:   "/foo?q=42",
		StatusCode:     http.StatusTeapot,
		ContentType:    "text/plain; charset=utf-8",
		StartedAt:      now,
		ProxyName:      "test:history",
		RequestEdited:  true,
		ResponseEdited: true,
	}, e.Metadata, t)

	tst.Is(false, strings.Contains(e.OriginalRequest(), "X-Request-Modified"), t)
	tst.Is(true, strings.Contains(e.EditedRequest(), "X-Request-Modified"), t)
	tst.Is(false, strings.Contains(e.OriginalResponse(), "X-Response-Modified"), t)
	tst.Is(true, strings.Contains(e.EditedResponse(), "X-Response-Modified"), t)

	tst.Is("true", hresp.Header.Get("X-Response-Modified"), t)
	tst.Is(evt.Metadata, e.Metadata, t)
	tst.Is(evt.Metadata,
		tst.DoB(r.Get(proxy.RoundTripID(evt.Metadata.ID)))(t).Metadata, t)
}

func TestMiddleWare(t *testing.T) {
	tst.Go(t)
	r := history.NewRecorder()
	var evt []history.Entry
	{
		evts := r.Events()
		defer r.Stop()
		go func() {
			for {
				select {
				case e := <-evts:
					evt = append(evt, e)
				case <-t.Context().Done():
				}
			}
		}()
	}

	const size = 100

	ca, caPool := proxytesting.SetupCert(t)
	p := tst.Do(proxy.New(ca, "test:history"))(t)
	sii := &stubInnerInterceptor{}
	p.Intercept(sii.mw())
	p.Intercept(r.MiddleWare())
	remote, cl := proxytesting.SetupProxyAndClient(t, caPool, p, nil)

	var wg sync.WaitGroup
	for i := range size {
		wg.Go(func() {
			str := strconv.Itoa(i)
			_ = tst.Do(cl.Post(remote+"/"+str, "text/plain", strings.NewReader(str)))(t)
		})
	}
	wg.Wait()

	got := r.GetAll()
	tst.Is(size, len(got), t)
	prev := got[0]
	for i, e := range got[1:] {
		if e.Metadata.ID > prev.Metadata.ID {
			prev = e
			continue
		}
		t.Errorf("Detected bad ordering: index %d entry has ID %d which is less than index %d with %d", i+1, e.Metadata.ID, i, prev.Metadata.ID)
	}

	gotUntil := r.GetUntil(size / 2)
	tst.Is(size/2, len(gotUntil), t)
}

package history_test

import (
	"context"
	"iter"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/empijei/chans"
	"github.com/empijei/tst"
	"github.com/empijei/web-proxy/history"
	"github.com/empijei/web-proxy/proxy"
	"github.com/empijei/web-proxy/testing/proxytesting"
	"github.com/empijei/web-proxy/ui"
)

func TestHistory(t *testing.T) {
	ctx := tst.Go(t)
	r := history.NewRecorder()
	const size = 15
	var i int

	ca, caPool := proxytesting.SetupCert(t)
	p := tst.Do(proxy.New(ca, "test:api_history"))(t)
	p.Intercept(r.MiddleWare())
	remote, cl := proxytesting.SetupProxyAndClient(t, caPool, p, nil)

	for range size / 3 {
		i++
		str := strconv.Itoa(i)
		_ = tst.Do(cl.Post(remote+"/"+str, "text/plain", strings.NewReader(str)))(t)
	}

	mux := http.NewServeMux()
	apiCtx, closeAPI := context.WithCancel(ctx)
	r.RegisterAPI(apiCtx, mux)
	apiSrv := httptest.NewServer(mux)

	rpc := ui.HistoryMetadataEP.RemoteWithOrigin(apiSrv.URL)
	req := ui.HistoryMetadataRequest{}

	first := tst.Do(rpc(ctx, req))(t)
	for range size / 3 {
		i++
		str := strconv.Itoa(i)
		_ = tst.Do(cl.Post(remote+"/"+str, "text/plain", strings.NewReader(str)))(t)
	}
	second := tst.Do(rpc(ctx, req))(t)
	for range size / 3 {
		i++
		str := strconv.Itoa(i)
		_ = tst.Do(cl.Post(remote+"/"+str, "text/plain", strings.NewReader(str)))(t)
	}
	third := tst.Do(rpc(ctx, req))(t)

	chans.Sleep(ctx.Done(), 50*time.Millisecond)
	closeAPI()

	consume := func(in iter.Seq2[ui.HistoryMetadataResponse, error]) []ui.HistoryMetadataResponse {
		t.Helper()
		var accum []ui.HistoryMetadataResponse
		for v, err := range in {
			tst.No(err, t)
			accum = append(accum, v)
		}
		return accum
	}
	got1 := consume(first)
	got2 := consume(second)
	got3 := consume(third)
	tst.Is(got1, got2, t)
	tst.Is(got2, got3, t)
	tst.Is(size, len(got1), t)
	// This test forces serialized data, so it should be returned in the same order.
	tst.Is(true, slices.IsSortedFunc(got1, func(a, b ui.HistoryMetadataResponse) int {
		return int(a.ID) - int(b.ID)
	}), t)
}

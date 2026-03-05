package ui

import (
	"net/http"
	"time"

	"github.com/empijei/srpc"
	"github.com/oklog/ulid/v2"
)

const (
	// ProxyPath is the path for all the proxy-related endpoints.
	ProxyPath = APIPath + "/proxy"
	// ProxyTrafficPath is the path to get traffic updates.
	ProxyTrafficPath = ProxyPath + "/traffic"
)

// ProxyTraffic is the endpoint to get traffic data from the proxy.
var ProxyTraffic = srpc.NewEndpoint(
	http.MethodGet,
	ProxyTrafficPath,
	srpc.NewCodecSeq[ProxyTrafficResponse](),
	srpc.NewCodecJSON[ProxyTrafficRequest](),
)

type (
	// ProxyTrafficRequest is a request to get proxy data.
	ProxyTrafficRequest struct {
		// Since allows clients to only ask for a subset of requests, based on time.
		Since time.Time
	}

	// TrafficOverview is a single entry for a proxy traffic response.
	TrafficOverview struct {
		ID           ulid.ULID
		Scheme       string
		Host         string
		Method       string
		PathAndQuery string
		StatusCode   int
		ContentType  string
		StartedAt    time.Time

		ProxyName      string
		RequestEdited  bool
		ResponseEdited bool
	}
	// ProxyTrafficResponse is a collection of events that happened since the last
	// update.
	ProxyTrafficResponse struct {
		Traffic []TrafficOverview
	}
)

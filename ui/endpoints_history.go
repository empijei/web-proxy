package ui

import (
	"net/http"
	"time"

	"github.com/empijei/srpc"
)

const (
	// HistoryPath is the path for all the proxy-related endpoints.
	HistoryPath = APIPath + "/history"
	// HistoryTrafficPath is the path to get traffic updates.
	HistoryTrafficPath = HistoryPath + "/traffic"
)

// HistoryMetadataEP is the endpoint to get live traffic metadata.
var HistoryMetadataEP = srpc.NewEndpointSeq[
	HistoryMetadataResponse,
	HistoryMetadataRequest](
	http.MethodGet,
	HistoryTrafficPath,
)

type (
	// HistoryMetadataRequest is a request to get proxy data.
	HistoryMetadataRequest struct{}

	// RoundTripID is the ID for a proxy roundtrip.
	RoundTripID uint64

	// HistoryMetadataResponse is a single entry for a proxy traffic response.
	HistoryMetadataResponse struct {
		ID           RoundTripID
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
)

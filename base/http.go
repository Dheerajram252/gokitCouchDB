package base

import (
	"context"
	"encoding/json"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var (
	healthCheckEndpoint = "/healthcheck"
	getDocument         = "/getDocument"
	putDocument         = "/putDocument"
)

func NewHTTPHandler(s Service, version, basePath string, transport TransportServerFinalizerInstrument) http.Handler {
	r := mux.NewRouter()
	{
		serverOptions := []httptransport.ServerOption{
			httptransport.ServerBefore(httptransport.PopulateRequestContext, setStartTime),
			httptransport.ServerErrorEncoder(encodeErrorMessage),
			httptransport.ServerFinalizer(transport.TransportServerFinalizer),
		}
		fmt.Println(serverOptions)
		e := NewServerEndPoints(s)
		baseRoute := fmt.Sprintf("/%s/%s", basePath, version)
		fmt.Println(baseRoute)
		r.Methods(http.MethodGet).Path(healthCheckEndpoint).Handler(httptransport.NewServer(
			e.Check,
			httptransport.NopRequestDecoder,
			encodeHealthCheckResponse,
		))
		r.Methods(http.MethodGet).Path(baseRoute + getDocument).Handler(httptransport.NewServer(
			e.GetDocument,
			decodeGetRequest,
			encodeHealthCheckResponse,
			serverOptions...,
		))
		r.Methods(http.MethodGet).Path(baseRoute + putDocument).Handler(httptransport.NewServer(
			e.PutDocument,
			httptransport.NopRequestDecoder,
			encodeHealthCheckResponse,
			serverOptions...,
		))
	}

	return r
}

func decodeGetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return r.URL.Query().Get("id"), nil
}

func encodeHealthCheckResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorMessage(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func setStartTime(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, contextStartTime, time.Now())
}
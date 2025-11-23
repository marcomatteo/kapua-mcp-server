package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"kapua-mcp-server/internal/kapua/config"
	"kapua-mcp-server/internal/kapua/models"
	"kapua-mcp-server/internal/kapua/services"
	"kapua-mcp-server/pkg/utils"
)

type handlerRoundTripper struct {
	handler http.Handler
}

func (rt handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	rt.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}

func newKapuaTestHandler(t *testing.T, handler http.HandlerFunc, loggerName string) *KapuaHandler {
	t.Helper()

	client := services.NewKapuaClient(&config.KapuaConfig{APIEndpoint: "http://kapua.test", Timeout: 5})
	client.SetHTTPClient(&http.Client{
		Transport: handlerRoundTripper{handler: handler},
	})
	client.SetTokenInfo(&models.AccessToken{KapuaEntity: models.KapuaEntity{ScopeID: models.KapuaID("tenant")}})

	return &KapuaHandler{client: client, logger: utils.NewDefaultLogger(loggerName)}
}

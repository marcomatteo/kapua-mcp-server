package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
)

type dataClientRoundTripFunc func(*http.Request) (*http.Response, error)

func (f dataClientRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestListDataClientsSuccess(t *testing.T) {
	client := newTestKapuaClient()

	params := map[string]string{
		"clientId": "Client-1",
		"limit":    "10",
		"offset":   "5",
	}

	sampleResp := `{"type":"clientInfoListResult","limitExceeded":false,"size":1,"totalCount":1,"items":[{"type":"clientInfo","id":"client-info-1","scopeId":"tenant","clientId":"Client-1","firstMessageId":"11111111-1111-1111-1111-111111111111","firstMessageOn":"2023-09-12T08:14:13.228Z","lastMessageId":"22222222-2222-2222-2222-222222222222","lastMessageOn":"2023-09-12T09:25:05.096Z"}]}`

	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/data/clients" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		if got := req.URL.Query().Get("clientId"); got != "Client-1" {
			t.Fatalf("expected clientId filter, got %q", got)
		}
		if got := req.URL.Query().Get("limit"); got != "10" {
			t.Fatalf("expected limit, got %q", got)
		}
		if got := req.URL.Query().Get("offset"); got != "5" {
			t.Fatalf("expected offset, got %q", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.ListDataClients(context.Background(), params)
	if err != nil {
		t.Fatalf("ListDataClients returned error: %v", err)
	}
	if result == nil || len(result.Items) != 1 {
		t.Fatalf("expected one data client, got %+v", result)
	}
	if result.Items[0].ClientID != "Client-1" {
		t.Fatalf("unexpected client info: %+v", result.Items[0])
	}
}

func TestListDataClientsRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := client.ListDataClients(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "list data clients request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestListDataClientsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"code":"500","message":"kapua error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ListDataClients(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "failed to list data clients") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestCountDataClientsSuccess(t *testing.T) {
	client := newTestKapuaClient()

	query := &models.KapuaQuery{Limit: 25, Offset: 5, AskTotalCount: true}

	sampleResp := `{"count":7}`

	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/data/clients/_count" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}
		if payload["limit"].(float64) != 25 || payload["offset"].(float64) != 5 {
			t.Fatalf("unexpected payload values: %v", payload)
		}
		if ask, ok := payload["askTotalCount"].(bool); !ok || !ask {
			t.Fatalf("expected askTotalCount true, got %v", payload["askTotalCount"])
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.CountDataClients(context.Background(), query)
	if err != nil {
		t.Fatalf("CountDataClients returned error: %v", err)
	}
	if result == nil || result.Count != 7 {
		t.Fatalf("unexpected count result: %+v", result)
	}
}

func TestCountDataClientsNilQuery(t *testing.T) {
	client := newTestKapuaClient()

	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if len(strings.TrimSpace(string(body))) != 2 { // {}
			t.Fatalf("expected empty query payload, got %q", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"count":0}`)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.CountDataClients(context.Background(), nil)
	if err != nil {
		t.Fatalf("CountDataClients returned error: %v", err)
	}
	if result.Count != 0 {
		t.Fatalf("expected count 0, got %d", result.Count)
	}
}

func TestCountDataClientsHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("kapua error")),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.CountDataClients(context.Background(), &models.KapuaQuery{})
	if err == nil || !strings.Contains(err.Error(), "failed to count data clients") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestGetDataClientSuccess(t *testing.T) {
	client := newTestKapuaClient()

	sampleResp := `{"type":"clientInfo","id":"client-info-1","scopeId":"tenant","clientId":"Client-1","firstMessageOn":"2023-09-12T08:14:13.228Z","lastMessageOn":"2023-09-12T09:25:05.096Z"}`

	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/data/clients/client-info-1" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.GetDataClient(context.Background(), "client-info-1")
	if err != nil {
		t.Fatalf("GetDataClient returned error: %v", err)
	}
	if result.ClientID != "Client-1" {
		t.Fatalf("unexpected client ID: %s", result.ClientID)
	}
	if result.FirstMessageOn.IsZero() || result.LastMessageOn.IsZero() {
		t.Fatalf("expected timestamps to be parsed, got %+v", result)
	}
}

func TestGetDataClientRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network failure")
	})}

	_, err := client.GetDataClient(context.Background(), "client-info-1")
	if err == nil || !strings.Contains(err.Error(), "get data client request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestGetDataClientHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataClientRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"code":"500","message":"kapua error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.GetDataClient(context.Background(), "client-info-1")
	if err == nil || !strings.Contains(err.Error(), "failed to get data client") {
		t.Fatalf("expected response error, got %v", err)
	}
}

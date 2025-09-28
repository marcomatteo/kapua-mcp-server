package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type dataMessageRoundTripFunc func(*http.Request) (*http.Response, error)

func (f dataMessageRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func intPtr(v int) *int {
	return &v
}

func TestListDataMessagesSuccess(t *testing.T) {
	client := newTestKapuaClient()

	query := &DataMessagesQuery{
		ClientIDs: []string{"B827EBFE9702", "AC40EA03D857", "esf_training_4411"},
		SortDir:   "DESC",
		Limit:     intPtr(50),
		Offset:    intPtr(0),
	}

	sampleResp := `{"type":"dataMessageListResult","limitExceeded":false,"size":1,"totalCount":1,"items":[{"type":"jsonDatastoreMessage","datastoreId":"6349cec8-396b-4aac-bc2f-8fca9fe0c67c","scopeId":"tenant","timestamp":"2023-09-12T09:35:04.383Z","deviceId":"WyczTs_GuDM","clientId":"B827EBFE9702","receivedOn":"2023-09-12T09:35:04.389Z","sentOn":"2023-09-12T09:35:04.383Z","capturedOn":"2023-09-12T09:35:04.383Z"}]}`

	client.httpClient = &http.Client{Transport: dataMessageRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", req.Method)
		}
		if req.URL.Path != "/v1/tenant/data/messages" {
			t.Fatalf("unexpected path %s", req.URL.Path)
		}

		values := req.URL.Query()["clientId"]
		expected := []string{"B827EBFE9702", "AC40EA03D857", "esf_training_4411"}
		if len(values) != len(expected) || !slices.Equal(values, expected) {
			t.Fatalf("unexpected clientId values: %v", values)
		}
		if got := req.URL.Query().Get("sortDir"); got != "DESC" {
			t.Fatalf("expected sortDir DESC, got %q", got)
		}
		if got := req.URL.Query().Get("limit"); got != "50" {
			t.Fatalf("expected limit 50, got %q", got)
		}
		if got := req.URL.Query().Get("offset"); got != "0" {
			t.Fatalf("expected offset 0, got %q", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(sampleResp)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.ListDataMessages(context.Background(), query)
	if err != nil {
		t.Fatalf("ListDataMessages returned error: %v", err)
	}
	if result == nil || len(result.Items) != 1 {
		t.Fatalf("expected one data message, got %+v", result)
	}
	if result.Items[0].ClientID != "B827EBFE9702" {
		t.Fatalf("unexpected data message: %+v", result.Items[0])
	}
}

func TestListDataMessagesNilQuery(t *testing.T) {
	client := newTestKapuaClient()

	client.httpClient = &http.Client{Transport: dataMessageRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.RawQuery != "" {
			t.Fatalf("expected no query parameters, got %q", req.URL.RawQuery)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"items":[]}`)),
			Header:     make(http.Header),
		}, nil
	})}

	result, err := client.ListDataMessages(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListDataMessages returned error: %v", err)
	}
	if result == nil || len(result.Items) != 0 {
		t.Fatalf("expected empty items, got %+v", result)
	}
}

func TestListDataMessagesRequestError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataMessageRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := client.ListDataMessages(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "list data messages request failed") {
		t.Fatalf("expected wrapped request error, got %v", err)
	}
}

func TestListDataMessagesHandleError(t *testing.T) {
	client := newTestKapuaClient()
	client.httpClient = &http.Client{Transport: dataMessageRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"code":"500","message":"kapua error"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := client.ListDataMessages(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "failed to list data messages") {
		t.Fatalf("expected response error, got %v", err)
	}
}

func TestDataMessagesQueryValues(t *testing.T) {
	strict := true
	limit := 25
	offset := 5
	query := &DataMessagesQuery{
		ClientIDs:     []string{"A", "B"},
		Channel:       "/foo/bar",
		StrictChannel: &strict,
		StartDate:     "2023-09-01T00:00:00Z",
		EndDate:       "2023-09-02T00:00:00Z",
		SortDir:       "ASC",
		Limit:         &limit,
		Offset:        &offset,
	}

	values := query.toValues()
	if len(values["clientId"]) != 2 {
		t.Fatalf("expected two clientId params, got %v", values["clientId"])
	}
	expectedPairs := map[string]string{
		"channel":   "/foo/bar",
		"startDate": "2023-09-01T00:00:00Z",
		"endDate":   "2023-09-02T00:00:00Z",
		"sortDir":   "ASC",
		"limit":     "25",
		"offset":    "5",
	}
	for key, value := range expectedPairs {
		if values.Get(key) != value {
			t.Fatalf("expected %s=%s, got %s", key, value, values.Get(key))
		}
	}
	if got := values.Get("strictChannel"); got != strconv.FormatBool(strict) {
		t.Fatalf("unexpected strictChannel value: %s", got)
	}
}

func TestDataMessagesQueryNil(t *testing.T) {
	var query *DataMessagesQuery
	values := query.toValues()
	if len(url.Values(values)) != 0 {
		t.Fatalf("expected empty values for nil query, got %v", values)
	}
}

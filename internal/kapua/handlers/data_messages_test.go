package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"kapua-mcp-server/internal/kapua/models"
)

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestHandleListDataMessagesSuccess(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/tenant/data/messages" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		values := r.URL.Query()["clientId"]
		expected := []string{"B827EBFE9702", "AC40EA03D857"}
		if len(values) != len(expected) || !slices.Equal(values, expected) {
			t.Fatalf("unexpected clientId values: %v", values)
		}
		if got := r.URL.Query().Get("strictChannel"); got != "true" {
			t.Fatalf("expected strictChannel true, got %s", got)
		}
		if got := r.URL.Query().Get("limit"); got != "50" {
			t.Fatalf("expected limit 50, got %s", got)
		}
		if got := r.URL.Query().Get("offset"); got != "0" {
			t.Fatalf("expected offset 0, got %s", got)
		}

		payload := models.DataMessageListResult{
			Items: []models.DataMessage{{DatastoreID: "message-1", ClientID: "B827EBFE9702"}},
			Size:  1,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	params := &ListDataMessagesParams{
		ClientIDs:     []string{"B827EBFE9702", "AC40EA03D857"},
		StrictChannel: boolPtr(true),
		Limit:         intPtr(50),
		Offset:        intPtr(0),
	}

	result, data, err := handler.HandleListDataMessages(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("HandleListDataMessages returned error: %v", err)
	}

	if len(result.Content) != 2 {
		t.Fatalf("expected two content entries, got %d", len(result.Content))
	}

	summary := textContent(t, result.Content[0])
	if summary != "Found 1 data messages." {
		t.Fatalf("unexpected summary: %s", summary)
	}

	var response models.DataMessageListResult
	if err := json.Unmarshal([]byte(textContent(t, result.Content[1])), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].DatastoreID != "message-1" {
		t.Fatalf("unexpected list payload: %+v", response)
	}

	typed, ok := data.(*models.DataMessageListResult)
	if !ok {
		t.Fatalf("expected *models.DataMessageListResult, got %T", data)
	}
	if len(typed.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(typed.Items))
	}
}

func TestHandleListDataMessagesNoParams(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Fatalf("expected no query parameters, got %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	})

	result, data, err := handler.HandleListDataMessages(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("HandleListDataMessages returned error: %v", err)
	}

	summary := textContent(t, result.Content[0])
	if summary != "Found 0 data messages." {
		t.Fatalf("unexpected summary: %s", summary)
	}

	typed, ok := data.(*models.DataMessageListResult)
	if !ok || len(typed.Items) != 0 {
		t.Fatalf("unexpected result: %+v", data)
	}
}

func TestHandleListDataMessagesServiceError(t *testing.T) {
	handler := newDeviceHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":"ERR"}`))
	})

	_, _, err := handler.HandleListDataMessages(context.Background(), nil, &ListDataMessagesParams{})
	if err == nil || !strings.Contains(err.Error(), "failed to list data messages") {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

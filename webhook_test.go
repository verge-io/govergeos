package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestWebhookURLService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhook_urls": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []WebhookURL{
				{Key: 1, Name: "slack", URL: "https://hooks.slack.com/test"},
				{Key: 2, Name: "pagerduty", URL: "https://events.pagerduty.com/test"},
			})
		},
	}))

	webhooks, err := client.WebhookURLs.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(webhooks) != 2 {
		t.Fatalf("expected 2 webhook URLs, got %d", len(webhooks))
	}
}

func TestWebhookURLService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhook_urls/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, WebhookURL{Key: 1, Name: "slack", URL: "https://hooks.slack.com/test"})
		},
	}))

	wh, err := client.WebhookURLs.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if wh.Name != "slack" {
		t.Errorf("expected name 'slack', got %q", wh.Name)
	}
}

func TestWebhookURLService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhook_urls": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []WebhookURL{{Key: 1, Name: "slack"}})
		},
		"GET /api/v4/webhook_urls/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, WebhookURL{Key: 1, Name: "slack", URL: "https://hooks.slack.com"})
		},
	}))

	wh, err := client.WebhookURLs.GetByName(context.Background(), "slack")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if wh.Name != "slack" {
		t.Errorf("expected name 'slack', got %q", wh.Name)
	}
}

func TestWebhookURLService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhook_urls": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []WebhookURL{})
		},
	}))

	_, err := client.WebhookURLs.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestWebhookURLService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/webhook_urls": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, apiResponse{Key: float64(1)})
		},
		"GET /api/v4/webhook_urls/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, WebhookURL{Key: 1, Name: "slack", URL: "https://hooks.slack.com"})
		},
	}))

	wh, err := client.WebhookURLs.Create(context.Background(), &WebhookURLCreateRequest{
		Name: "slack",
		URL:  "https://hooks.slack.com",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if wh.Name != "slack" {
		t.Errorf("expected name 'slack', got %q", wh.Name)
	}
}

func TestWebhookURLService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.WebhookURLs.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestWebhookURLService_Create_MissingName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.WebhookURLs.Create(context.Background(), &WebhookURLCreateRequest{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestWebhookURLService_Create_MissingURL(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.WebhookURLs.Create(context.Background(), &WebhookURLCreateRequest{Name: "test"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestWebhookURLService_Update(t *testing.T) {
	newName := "updated-slack"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/webhook_urls/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
		"GET /api/v4/webhook_urls/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, WebhookURL{Key: 1, Name: newName})
		},
	}))

	wh, err := client.WebhookURLs.Update(context.Background(), 1, &WebhookURLUpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if wh.Name != newName {
		t.Errorf("expected name %q, got %q", newName, wh.Name)
	}
}

func TestWebhookURLService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.WebhookURLs.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWebhookURLService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/webhook_urls/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.WebhookURLs.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestWebhookURLService_Send(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/webhook_url_actions": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			if body["action"] != "send" {
				t.Errorf("expected action 'send', got %v", body["action"])
			}
			w.WriteHeader(200)
		},
	}))

	err := client.WebhookURLs.Send(context.Background(), 1, "test message")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

// Webhook (message log) tests

func TestWebhookService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhooks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Webhook{
				{Key: 1, Status: "sent", Message: "test"},
				{Key: 2, Status: "queued", Message: "test2"},
			})
		},
	}))

	webhooks, err := client.Webhooks.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(webhooks) != 2 {
		t.Fatalf("expected 2 webhooks, got %d", len(webhooks))
	}
}

func TestWebhookService_ListByWebhookURL(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhooks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Webhook{{Key: 1, WebhookURL: 5}})
		},
	}))

	webhooks, err := client.Webhooks.ListByWebhookURL(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListByWebhookURL failed: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(webhooks))
	}
}

func TestWebhookService_ListByStatus(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhooks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Webhook{{Key: 1, Status: "error"}})
		},
	}))

	webhooks, err := client.Webhooks.ListByStatus(context.Background(), "error")
	if err != nil {
		t.Fatalf("ListByStatus failed: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(webhooks))
	}
}

func TestWebhookService_ListPending(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhooks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Webhook{{Key: 1, Status: "queued"}})
		},
	}))

	webhooks, err := client.Webhooks.ListPending(context.Background())
	if err != nil {
		t.Fatalf("ListPending failed: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(webhooks))
	}
}

func TestWebhookService_ListFailed(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhooks": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Webhook{{Key: 1, Status: "error"}})
		},
	}))

	webhooks, err := client.Webhooks.ListFailed(context.Background())
	if err != nil {
		t.Fatalf("ListFailed failed: %v", err)
	}
	if len(webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(webhooks))
	}
}

func TestWebhookService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/webhooks/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Webhook{Key: 1, Status: "sent", Message: "payload"})
		},
	}))

	wh, err := client.Webhooks.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if wh.Status != "sent" {
		t.Errorf("expected status 'sent', got %q", wh.Status)
	}
}

func TestWebhookService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/webhooks/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Webhooks.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

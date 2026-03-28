package vergeos

import (
	"context"
	"net/http"
	"testing"
)

func TestSystemService_GetInfo(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /version.json": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SystemInfo{Name: "v4", Version: "4.13.1", Hash: "abc123"})
		},
	}))

	info, err := client.System.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}
	if info.Version != "4.13.1" {
		t.Errorf("expected version '4.13.1', got %q", info.Version)
	}
	if info.Name != "v4" {
		t.Errorf("expected name 'v4', got %q", info.Name)
	}
	if info.Hash != "abc123" {
		t.Errorf("expected hash 'abc123', got %q", info.Hash)
	}
}

func TestSystemService_GetVersion(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /version.json": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, SystemInfo{Name: "v4", Version: "4.13.1"})
		},
	}))

	version, err := client.System.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}
	if version != "4.13.1" {
		t.Errorf("expected '4.13.1', got %q", version)
	}
}

func TestSystemService_GetInfo_ServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /version.json": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			w.Write([]byte("not json"))
		},
	}))

	_, err := client.System.GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid response")
	}
}

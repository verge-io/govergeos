package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestClusterService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Cluster{
				{Key: 1, Name: "cluster-a", Enabled: true},
				{Key: 2, Name: "cluster-b", Enabled: false},
			})
		},
	}))

	clusters, err := client.Clusters.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	if clusters[0].Name != "cluster-a" {
		t.Errorf("expected name 'cluster-a', got %q", clusters[0].Name)
	}
}

func TestClusterService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Cluster{Key: 1, Name: "cluster-a", Enabled: true, Description: "primary"})
		},
	}))

	cluster, err := client.Clusters.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cluster.Name != "cluster-a" {
		t.Errorf("expected name 'cluster-a', got %q", cluster.Name)
	}
	if cluster.Description != "primary" {
		t.Errorf("expected description 'primary', got %q", cluster.Description)
	}
}

func TestClusterService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Clusters.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterService_GetByName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Cluster{{Key: 1, Name: "cluster-a"}})
		},
	}))

	cluster, err := client.Clusters.GetByName(context.Background(), "cluster-a")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if cluster.Name != "cluster-a" {
		t.Errorf("expected name 'cluster-a', got %q", cluster.Name)
	}
}

func TestClusterService_GetByName_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Cluster{})
		},
	}))

	_, err := client.Clusters.GetByName(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterService_GetStatus(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, struct {
				Status ClusterStatus `json:"status"`
			}{
				Status: ClusterStatus{
					Cluster:    1,
					Status:     "online",
					TotalNodes: 3,
					OnlineNodes: 3,
				},
			})
		},
	}))

	status, err := client.Clusters.GetStatus(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status.Status != "online" {
		t.Errorf("expected status 'online', got %q", status.Status)
	}
	if status.TotalNodes != 3 {
		t.Errorf("expected 3 total nodes, got %d", status.TotalNodes)
	}
}

func TestClusterService_GetStatus_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/clusters/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Clusters.GetStatus(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/clusters": func(w http.ResponseWriter, r *http.Request) {
			var req ClusterCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "new-cluster" {
				t.Errorf("expected name 'new-cluster', got %q", req.Name)
			}
			if req.Enabled == nil || !*req.Enabled {
				t.Error("expected enabled to default to true")
			}
			jsonResponse(w, 200, apiResponse{Key: 10})
		},
		"GET /api/v4/clusters/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Cluster{Key: 10, Name: "new-cluster", Enabled: true})
		},
	}))

	cluster, err := client.Clusters.Create(context.Background(), &ClusterCreateRequest{
		Name: "new-cluster",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if cluster.Name != "new-cluster" {
		t.Errorf("expected name 'new-cluster', got %q", cluster.Name)
	}
}

func TestClusterService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Clusters.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestClusterService_Create_EmptyName(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Clusters.Create(context.Background(), &ClusterCreateRequest{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestClusterService_Update(t *testing.T) {
	newName := "renamed"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/clusters/1": func(w http.ResponseWriter, r *http.Request) {
			var req ClusterUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name == nil || *req.Name != newName {
				t.Errorf("expected name %q, got %v", newName, req.Name)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/clusters/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Cluster{Key: 1, Name: newName})
		},
	}))

	cluster, err := client.Clusters.Update(context.Background(), 1, &ClusterUpdateRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if cluster.Name != newName {
		t.Errorf("expected name %q, got %q", newName, cluster.Name)
	}
}

func TestClusterService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Clusters.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestClusterService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/clusters/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newName := "nope"
	_, err := client.Clusters.Update(context.Background(), 999, &ClusterUpdateRequest{Name: &newName})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestClusterService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/clusters/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Clusters.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestClusterService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/clusters/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Cluster.Delete returns nil for 404 (already deleted)
	err := client.Clusters.Delete(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected nil for not-found delete, got: %v", err)
	}
}

package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestMemberService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/members": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Member{
				{ID: 1, Group: 10, Member: "users/admin"},
				{ID: 2, Group: 10, Member: "users/operator"},
			})
		},
	}))

	members, err := client.Members.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].Member != "users/admin" {
		t.Errorf("expected member 'users/admin', got %q", members[0].Member)
	}
}

func TestMemberService_ListByGroup(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/members": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "parent_group eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []Member{
				{ID: 1, Group: 10, Member: "users/admin"},
			})
		},
	}))

	members, err := client.Members.ListByGroup(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByGroup failed: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
}

func TestMemberService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/members/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Member{ID: 1, Group: 10, Member: "users/admin"})
		},
	}))

	member, err := client.Members.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if member.Member != "users/admin" {
		t.Errorf("expected member 'users/admin', got %q", member.Member)
	}
	if int(member.Group) != 10 {
		t.Errorf("expected group 10, got %d", int(member.Group))
	}
}

func TestMemberService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/members/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.Members.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMemberService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/members": func(w http.ResponseWriter, r *http.Request) {
			var req MemberCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Group != 10 {
				t.Errorf("expected group 10, got %d", req.Group)
			}
			if req.Member != "users/newuser" {
				t.Errorf("expected member 'users/newuser', got %q", req.Member)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(3)})
		},
		"GET /api/v4/members/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Member{ID: 3, Group: 10, Member: "users/newuser"})
		},
	}))

	member, err := client.Members.Create(context.Background(), &MemberCreateRequest{
		Group:  10,
		Member: "users/newuser",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if member.Member != "users/newuser" {
		t.Errorf("expected member 'users/newuser', got %q", member.Member)
	}
	if int(member.ID) != 3 {
		t.Errorf("expected ID 3, got %d", int(member.ID))
	}
}

func TestMemberService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Members.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestMemberService_Create_MissingGroup(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Members.Create(context.Background(), &MemberCreateRequest{
		Member: "users/admin",
	})
	if err == nil {
		t.Fatal("expected error for missing group")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestMemberService_Create_MissingMember(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Members.Create(context.Background(), &MemberCreateRequest{
		Group: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing member")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestMemberService_Update(t *testing.T) {
	newMember := "users/updated"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/members/1": func(w http.ResponseWriter, r *http.Request) {
			var req MemberUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Member == nil || *req.Member != newMember {
				t.Errorf("expected member %q, got %v", newMember, req.Member)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/members/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Member{ID: 1, Group: 10, Member: newMember})
		},
	}))

	member, err := client.Members.Update(context.Background(), 1, &MemberUpdateRequest{Member: &newMember})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if member.Member != newMember {
		t.Errorf("expected member %q, got %q", newMember, member.Member)
	}
}

func TestMemberService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.Members.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestMemberService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/members/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newMember := "users/updated"
	_, err := client.Members.Update(context.Background(), 999, &MemberUpdateRequest{Member: &newMember})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestMemberService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/members/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Members.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestMemberService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/members/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete of already-deleted resource returns nil
	err := client.Members.Delete(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected nil for already-deleted member, got %v", err)
	}
}

func TestMemberService_Add(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/members": func(w http.ResponseWriter, r *http.Request) {
			var req MemberCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Group != 10 {
				t.Errorf("expected group 10, got %d", req.Group)
			}
			if req.Member != "users/admin" {
				t.Errorf("expected member 'users/admin', got %q", req.Member)
			}
			jsonResponse(w, 200, apiResponse{Key: float64(5)})
		},
		"GET /api/v4/members/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Member{ID: 5, Group: 10, Member: "users/admin"})
		},
	}))

	member, err := client.Members.Add(context.Background(), 10, "users/admin")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if member.Member != "users/admin" {
		t.Errorf("expected member 'users/admin', got %q", member.Member)
	}
}

func TestMemberService_Remove(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/members": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Member{
				{ID: 5, Group: 10, Member: "users/admin"},
			})
		},
		"DELETE /api/v4/members/5": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.Members.Remove(context.Background(), 10, "users/admin")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
}

func TestMemberService_Remove_AlreadyRemoved(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/members": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []Member{})
		},
	}))

	// Remove of non-existent membership returns nil
	err := client.Members.Remove(context.Background(), 10, "users/nonexistent")
	if err != nil {
		t.Fatalf("expected nil for already-removed member, got %v", err)
	}
}

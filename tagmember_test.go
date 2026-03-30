package vergeos

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTagMemberService_List(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TagMember{
				{Key: 1, Tag: 10, Member: "vms/123"},
				{Key: 2, Tag: 10, Member: "vms/456"},
			})
		},
	}))

	members, err := client.TagMembers.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].Member != "vms/123" {
		t.Errorf("expected member 'vms/123', got %q", members[0].Member)
	}
}

func TestTagMemberService_ListByTag(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "tag eq 10" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TagMember{{Key: 1, Tag: 10, Member: "vms/123"}})
		},
	}))

	members, err := client.TagMembers.ListByTag(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListByTag failed: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
}

func TestTagMemberService_ListByMember(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			filter := r.URL.Query().Get("filter")
			if filter != "member eq 'vms/123'" {
				t.Errorf("unexpected filter: %s", filter)
			}
			jsonResponse(w, 200, []TagMember{{Key: 1, Tag: 10, Member: "vms/123"}})
		},
	}))

	members, err := client.TagMembers.ListByMember(context.Background(), "vms/123")
	if err != nil {
		t.Fatalf("ListByMember failed: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
}

func TestTagMemberService_Get(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagMember{Key: 1, Tag: 10, Member: "vms/123"})
		},
	}))

	member, err := client.TagMembers.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if member.Member != "vms/123" {
		t.Errorf("expected member 'vms/123', got %q", member.Member)
	}
}

func TestTagMemberService_Get_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	_, err := client.TagMembers.Get(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Create(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			var req TagMemberCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Tag != 10 {
				t.Errorf("expected tag 10, got %d", req.Tag)
			}
			if req.Member != "vms/123" {
				t.Errorf("expected member 'vms/123', got %q", req.Member)
			}
			jsonResponse(w, 200, map[string]any{"$key": 1})
		},
		"GET /api/v4/tag_members/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagMember{Key: 1, Tag: 10, Member: "vms/123"})
		},
	}))

	member, err := client.TagMembers.Create(context.Background(), &TagMemberCreateRequest{
		Tag:    10,
		Member: "vms/123",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if member.Member != "vms/123" {
		t.Errorf("expected member 'vms/123', got %q", member.Member)
	}
}

func TestTagMemberService_Create_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagMembers.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Create_MissingTag(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagMembers.Create(context.Background(), &TagMemberCreateRequest{
		Member: "vms/123",
	})
	if err == nil {
		t.Fatal("expected error for missing tag")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Create_MissingMember(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagMembers.Create(context.Background(), &TagMemberCreateRequest{
		Tag: 10,
	})
	if err == nil {
		t.Fatal("expected error for missing member")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Create_InvalidMemberFormat(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagMembers.Create(context.Background(), &TagMemberCreateRequest{
		Tag:    10,
		Member: "invalid-no-slash",
	})
	if err == nil {
		t.Fatal("expected error for invalid member format")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Update(t *testing.T) {
	newMember := "vms/456"
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tag_members/1": func(w http.ResponseWriter, r *http.Request) {
			var req TagMemberUpdateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Member == nil || *req.Member != newMember {
				t.Errorf("expected member %q, got %v", newMember, req.Member)
			}
			w.WriteHeader(200)
		},
		"GET /api/v4/tag_members/1": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagMember{Key: 1, Tag: 10, Member: newMember})
		},
	}))

	member, err := client.TagMembers.Update(context.Background(), 1, &TagMemberUpdateRequest{Member: &newMember})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if member.Member != newMember {
		t.Errorf("expected member %q, got %q", newMember, member.Member)
	}
}

func TestTagMemberService_Update_NilRequest(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	_, err := client.TagMembers.Update(context.Background(), 1, nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Update_InvalidMemberFormat(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))

	badMember := "invalid-no-slash"
	_, err := client.TagMembers.Update(context.Background(), 1, &TagMemberUpdateRequest{Member: &badMember})
	if err == nil {
		t.Fatal("expected error for invalid member format")
	}
	if !IsValidationError(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Update_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/tag_members/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	newMember := "vms/456"
	_, err := client.TagMembers.Update(context.Background(), 999, &TagMemberUpdateRequest{Member: &newMember})
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Delete(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tag_members/1": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TagMembers.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestTagMemberService_Delete_NotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"DELETE /api/v4/tag_members/999": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 404, map[string]string{"err": "not found"})
		},
	}))

	// Delete returns nil for not found (already deleted)
	err := client.TagMembers.Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected NotFoundError for deleted tag member")
	}
	if !IsNotFoundError(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestTagMemberService_Assign(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			var req TagMemberCreateRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Tag != 10 {
				t.Errorf("expected tag 10, got %d", req.Tag)
			}
			if req.Member != "vms/123" {
				t.Errorf("expected member 'vms/123', got %q", req.Member)
			}
			jsonResponse(w, 200, map[string]any{"$key": 5})
		},
		"GET /api/v4/tag_members/5": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, TagMember{Key: 5, Tag: 10, Member: "vms/123"})
		},
	}))

	member, err := client.TagMembers.Assign(context.Background(), 10, "vms/123")
	if err != nil {
		t.Fatalf("Assign failed: %v", err)
	}
	if member.Tag.Int() != 10 {
		t.Errorf("expected tag 10, got %d", member.Tag.Int())
	}
	if member.Member != "vms/123" {
		t.Errorf("expected member 'vms/123', got %q", member.Member)
	}
}

func TestTagMemberService_Unassign(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TagMember{{Key: 5, Tag: 10, Member: "vms/123"}})
		},
		"DELETE /api/v4/tag_members/5": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		},
	}))

	err := client.TagMembers.Unassign(context.Background(), 10, "vms/123")
	if err != nil {
		t.Fatalf("Unassign failed: %v", err)
	}
}

func TestTagMemberService_Unassign_AlreadyUnassigned(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/tag_members": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []TagMember{})
		},
	}))

	// Should not error when already unassigned
	err := client.TagMembers.Unassign(context.Background(), 10, "vms/123")
	if err != nil {
		t.Fatalf("Unassign should not error when already unassigned, got: %v", err)
	}
}

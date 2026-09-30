package vergeos

import (
	"context"
	"net/http"
	"testing"
)

// Names that the old SQL-style escaper mishandled, and the filter literal
// VergeOS expects for each. \ is escaped before ' and {.
var filterNameCases = []struct {
	label  string
	name   string
	quoted string
}{
	{"apostrophe", "o'brien", `o\'brien`},
	{"braces", "zzgo-grp{x}", `zzgo-grp\{x}`},
	{"backslash", `back\slash`, `back\\slash`},
	{"combination", `o'brien{x}\z`, `o\'brien\{x}\\z`},
}

func TestGroupService_GetByName_EscapesAndMatchesExactly(t *testing.T) {
	for _, tt := range filterNameCases {
		t.Run(tt.label, func(t *testing.T) {
			wantFilter := "name eq '" + tt.quoted + "'"

			t.Run("query", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []Group{{Key: 2, Name: tt.name}})
					},
				}))

				group, err := client.Groups.GetByName(context.Background(), tt.name)
				if err != nil {
					t.Fatalf("GetByName: %v", err)
				}
				if group.Name != tt.name || group.Key != 2 {
					t.Fatalf("got %+v", group)
				}
			})

			t.Run("mismatch", func(t *testing.T) {
				// The platform consumed a brace group and returned a different group.
				returned := "zzgo-grp"
				if returned == tt.name {
					returned = "other"
				}
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/groups": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []Group{{Key: 2, Name: returned}})
					},
				}))

				_, err := client.Groups.GetByName(context.Background(), tt.name)
				notFound, ok := err.(*NotFoundError)
				if !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
				if notFound.Resource != "Group" || notFound.ID != tt.name {
					t.Fatalf("unexpected not-found details: %+v", notFound)
				}
			})
		})
	}
}

func TestVolumeService_GetByName_EscapesAndMatchesExactly(t *testing.T) {
	for _, tt := range filterNameCases {
		t.Run(tt.label, func(t *testing.T) {
			wantFilter := "service eq 5 and name eq '" + tt.quoted + "'"

			t.Run("query", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []Volume{{Key: "abc", ID: "abc", Name: tt.name}})
					},
					"GET /api/v4/volumes/abc": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, Volume{Key: "abc", ID: "abc", Name: tt.name})
					},
				}))

				vol, err := client.Volumes.GetByName(context.Background(), 5, tt.name)
				if err != nil {
					t.Fatalf("GetByName: %v", err)
				}
				if vol.Name != tt.name {
					t.Fatalf("got name %q", vol.Name)
				}
			})

			t.Run("list mismatch", func(t *testing.T) {
				returned := "zzgo-grp"
				if returned == tt.name {
					returned = "other"
				}
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []Volume{{Key: "abc", ID: "abc", Name: returned}})
					},
					"GET /api/v4/volumes/abc": func(w http.ResponseWriter, r *http.Request) {
						t.Error("Get was called for a list row whose name did not match")
						jsonResponse(w, 200, Volume{Key: "abc", ID: "abc", Name: returned})
					},
				}))

				_, err := client.Volumes.GetByName(context.Background(), 5, tt.name)
				if _, ok := err.(*NotFoundError); !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
			})

			t.Run("fetched name mismatch", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/volumes": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, []Volume{{Key: "abc", ID: "abc", Name: tt.name}})
					},
					"GET /api/v4/volumes/abc": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, Volume{Key: "abc", ID: "abc", Name: "other-volume"})
					},
				}))

				_, err := client.Volumes.GetByName(context.Background(), 5, tt.name)
				notFound, ok := err.(*NotFoundError)
				if !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
				if notFound.Resource != "Volume" || notFound.ID != tt.name {
					t.Fatalf("unexpected not-found details: %+v", notFound)
				}
			})
		})
	}
}

func TestVMService_GetByName_EscapesAndMatchesExactly(t *testing.T) {
	for _, tt := range filterNameCases {
		t.Run(tt.label, func(t *testing.T) {
			wantFilter := "is_snapshot eq false and name eq '" + tt.quoted + "'"

			t.Run("query", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []VM{{Key: 2, Name: tt.name}})
					},
					"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, VM{Key: 2, Name: tt.name})
					},
				}))

				vm, err := client.VMs.GetByName(context.Background(), tt.name)
				if err != nil {
					t.Fatalf("GetByName: %v", err)
				}
				if vm.Name != tt.name || vm.Key != 2 {
					t.Fatalf("got %+v", vm)
				}
			})

			t.Run("list mismatch", func(t *testing.T) {
				returned := "zzgo-grp"
				if returned == tt.name {
					returned = "other"
				}
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []VM{{Key: 2, Name: returned}})
					},
					"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
						t.Error("Get was called for a list row whose name did not match")
						jsonResponse(w, 200, VM{Key: 2, Name: returned})
					},
				}))

				_, err := client.VMs.GetByName(context.Background(), tt.name)
				notFound, ok := err.(*NotFoundError)
				if !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
				if notFound.Resource != "VM" || notFound.ID != tt.name {
					t.Fatalf("unexpected not-found details: %+v", notFound)
				}
			})

			t.Run("fetched name mismatch", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/vms": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, []VM{{Key: 2, Name: tt.name}})
					},
					"GET /api/v4/vms/2": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, VM{Key: 2, Name: "other-vm"})
					},
				}))

				_, err := client.VMs.GetByName(context.Background(), tt.name)
				notFound, ok := err.(*NotFoundError)
				if !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
				if notFound.Resource != "VM" || notFound.ID != tt.name {
					t.Fatalf("unexpected not-found details: %+v", notFound)
				}
			})
		})
	}
}

func TestNetworkService_GetByName_EscapesAndMatchesExactly(t *testing.T) {
	for _, tt := range filterNameCases {
		t.Run(tt.label, func(t *testing.T) {
			wantFilter := "name eq '" + tt.quoted + "'"

			t.Run("query", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []Network{{Key: 3, Name: tt.name}})
					},
					"GET /api/v4/vnets/3": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, Network{Key: 3, Name: tt.name})
					},
				}))

				network, err := client.Networks.GetByName(context.Background(), tt.name)
				if err != nil {
					t.Fatalf("GetByName: %v", err)
				}
				if network.Name != tt.name || network.Key != 3 {
					t.Fatalf("got %+v", network)
				}
			})

			t.Run("list mismatch", func(t *testing.T) {
				returned := "zzgo-grp"
				if returned == tt.name {
					returned = "other"
				}
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
						if got := r.URL.Query().Get("filter"); got != wantFilter {
							t.Errorf("filter = %q, want %q", got, wantFilter)
						}
						jsonResponse(w, 200, []Network{{Key: 3, Name: returned}})
					},
					"GET /api/v4/vnets/3": func(w http.ResponseWriter, r *http.Request) {
						t.Error("Get was called for a list row whose name did not match")
						jsonResponse(w, 200, Network{Key: 3, Name: returned})
					},
				}))

				_, err := client.Networks.GetByName(context.Background(), tt.name)
				notFound, ok := err.(*NotFoundError)
				if !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
				if notFound.Resource != "Network" || notFound.ID != tt.name {
					t.Fatalf("unexpected not-found details: %+v", notFound)
				}
			})

			t.Run("fetched name mismatch", func(t *testing.T) {
				client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
					"GET /api/v4/vnets": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, []Network{{Key: 3, Name: tt.name}})
					},
					"GET /api/v4/vnets/3": func(w http.ResponseWriter, r *http.Request) {
						jsonResponse(w, 200, Network{Key: 3, Name: "other-network"})
					},
				}))

				_, err := client.Networks.GetByName(context.Background(), tt.name)
				notFound, ok := err.(*NotFoundError)
				if !ok {
					t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
				}
				if notFound.Resource != "Network" || notFound.ID != tt.name {
					t.Fatalf("unexpected not-found details: %+v", notFound)
				}
			})
		})
	}
}

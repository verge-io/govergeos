package vergeos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func boolPtr(v bool) *bool { return &v }

func strPtr(v string) *string { return &v }

func decodeJSONBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

func assertJSONBody(t *testing.T, body map[string]any, want map[string]any, absent []string) {
	t.Helper()
	for key, value := range want {
		if body[key] != value {
			t.Errorf("body[%s] = %#v, want %#v (full %#v)", key, body[key], value, body)
		}
	}
	for _, key := range absent {
		if _, ok := body[key]; ok {
			t.Errorf("body[%s] should be omitted, got %#v", key, body[key])
		}
	}
}

func routingParentRoutes(needRestart bool) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bgp/4": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBGP{Key: 4, VNet: 10})
		},
		"GET /api/v4/vnet_bgp_routers/8": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBGPRouter{Key: 8, BGP: 4, ASN: 65000})
		},
		"GET /api/v4/vnet_bgp_interfaces/6": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBGPInterface{Key: 6, BGP: 4, Name: "uplink"})
		},
		"GET /api/v4/vnet_bgp_routemaps/3": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBGPRouteMap{Key: 3, BGP: 4, Tag: "IMPORT", Sequence: 10, Permit: true})
		},
		"GET /api/v4/vnet_eigrp_routers/2": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetEIGRPRouter{Key: 2, BGP: 4, ASN: 100})
		},
		"GET /api/v4/vnets/10": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, Network{Key: 10, NeedRestart: needRestart})
		},
	}
}

func mergeRoutes(parts ...map[string]http.HandlerFunc) map[string]http.HandlerFunc {
	out := map[string]http.HandlerFunc{}
	for _, part := range parts {
		for key, handler := range part {
			out[key] = handler
		}
	}
	return out
}

func assertPendingRestart(t *testing.T, status *RoutingRestartStatus) {
	t.Helper()
	if status == nil || status.NetworkID != 10 || !status.Pending || status.Restarted {
		t.Fatalf("unexpected restart status: %+v", status)
	}
}

func TestRouting_ListQueries(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		fields string
		sort   string
		filter string
		call   func(*Client) error
	}{
		{
			name:   "bgp",
			path:   "/api/v4/vnet_bgp",
			fields: vnetBGPListFields,
			call: func(c *Client) error {
				_, err := c.VNetBGP.List(context.Background())
				return err
			},
		},
		{
			name:   "bgp by network",
			path:   "/api/v4/vnet_bgp",
			fields: vnetBGPListFields,
			filter: "vnet eq 10",
			call: func(c *Client) error {
				_, err := c.VNetBGP.ListByNetwork(context.Background(), 10)
				return err
			},
		},
		{
			name:   "bgp routers",
			path:   "/api/v4/vnet_bgp_routers",
			fields: vnetBGPRouterListFields,
			sort:   "asn",
			filter: "bgp eq 4 and asn eq 65000",
			call: func(c *Client) error {
				_, err := c.VNetBGPRouters.ListByBGP(context.Background(), 4, WithFilter("asn eq 65000"))
				return err
			},
		},
		{
			name:   "bgp router sort override",
			path:   "/api/v4/vnet_bgp_routers",
			fields: "all",
			sort:   "-asn",
			call: func(c *Client) error {
				_, err := c.VNetBGPRouters.List(context.Background(), WithFields("all"), WithSort("-asn"))
				return err
			},
		},
		{
			name:   "bgp router commands",
			path:   "/api/v4/vnet_bgp_router_commands",
			fields: vnetBGPRouterCommandListFields,
			sort:   "orderid",
			filter: "bgp_router eq 8",
			call: func(c *Client) error {
				_, err := c.VNetBGPRouterCommands.ListByRouter(context.Background(), 8)
				return err
			},
		},
		{
			name:   "bgp interfaces",
			path:   "/api/v4/vnet_bgp_interfaces",
			fields: vnetBGPInterfaceListFields,
			sort:   "name",
			filter: "bgp eq 4",
			call: func(c *Client) error {
				_, err := c.VNetBGPInterfaces.ListByBGP(context.Background(), 4)
				return err
			},
		},
		{
			name:   "bgp interface commands",
			path:   "/api/v4/vnet_bgp_interface_commands",
			fields: vnetBGPInterfaceCommandListFields,
			sort:   "orderid",
			filter: "bgp_interface eq 6",
			call: func(c *Client) error {
				_, err := c.VNetBGPInterfaceCommands.ListByInterface(context.Background(), 6)
				return err
			},
		},
		{
			name:   "route maps",
			path:   "/api/v4/vnet_bgp_routemaps",
			fields: vnetBGPRouteMapListFields,
			sort:   "tag,sequence",
			filter: "bgp eq 4",
			call: func(c *Client) error {
				_, err := c.VNetBGPRouteMaps.ListByBGP(context.Background(), 4)
				return err
			},
		},
		{
			name:   "route map commands",
			path:   "/api/v4/vnet_bgp_routemap_commands",
			fields: vnetBGPRouteMapCommandListFields,
			sort:   "orderid",
			filter: "bgp_routemap eq 3",
			call: func(c *Client) error {
				_, err := c.VNetBGPRouteMapCommands.ListByRouteMap(context.Background(), 3)
				return err
			},
		},
		{
			name:   "ip commands",
			path:   "/api/v4/vnet_bgp_ip",
			fields: vnetBGPIPCommandListFields,
			sort:   "orderid",
			filter: "bgp eq 4",
			call: func(c *Client) error {
				_, err := c.VNetBGPIPCommands.ListByBGP(context.Background(), 4)
				return err
			},
		},
		{
			name:   "ospf commands",
			path:   "/api/v4/vnet_ospf_commands",
			fields: vnetOSPFCommandListFields,
			sort:   "orderid",
			filter: "bgp eq 4",
			call: func(c *Client) error {
				_, err := c.VNetOSPFCommands.ListByBGP(context.Background(), 4)
				return err
			},
		},
		{
			name:   "eigrp routers",
			path:   "/api/v4/vnet_eigrp_routers",
			fields: vnetEIGRPRouterListFields,
			sort:   "asn",
			filter: "bgp eq 4",
			call: func(c *Client) error {
				_, err := c.VNetEIGRPRouters.ListByBGP(context.Background(), 4)
				return err
			},
		},
		{
			name:   "eigrp router commands",
			path:   "/api/v4/vnet_eigrp_router_commands",
			fields: vnetEIGRPRouterCommandListFields,
			sort:   "orderid",
			filter: "eigrp_router eq 2",
			call: func(c *Client) error {
				_, err := c.VNetEIGRPRouterCommands.ListByRouter(context.Background(), 2)
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"GET " + tc.path: func(w http.ResponseWriter, r *http.Request) {
					if got := r.URL.Query().Get("fields"); got != tc.fields {
						t.Errorf("fields = %q, want %q", got, tc.fields)
					}
					if got := r.URL.Query().Get("sort"); got != tc.sort {
						t.Errorf("sort = %q, want %q", got, tc.sort)
					}
					if got := r.URL.Query().Get("filter"); got != tc.filter {
						t.Errorf("filter = %q, want %q", got, tc.filter)
					}
					jsonResponse(w, 200, []map[string]any{})
				},
			}))
			if err := tc.call(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRouting_ListServerError(t *testing.T) {
	cases := []struct {
		name string
		path string
		call func(*Client) error
	}{
		{name: "bgp", path: "/api/v4/vnet_bgp", call: func(c *Client) error {
			_, err := c.VNetBGP.List(context.Background())
			return err
		}},
		{name: "routers", path: "/api/v4/vnet_bgp_routers", call: func(c *Client) error {
			_, err := c.VNetBGPRouters.List(context.Background())
			return err
		}},
		{name: "interfaces", path: "/api/v4/vnet_bgp_interfaces", call: func(c *Client) error {
			_, err := c.VNetBGPInterfaces.List(context.Background())
			return err
		}},
		{name: "route maps", path: "/api/v4/vnet_bgp_routemaps", call: func(c *Client) error {
			_, err := c.VNetBGPRouteMaps.List(context.Background())
			return err
		}},
		{name: "ip", path: "/api/v4/vnet_bgp_ip", call: func(c *Client) error {
			_, err := c.VNetBGPIPCommands.List(context.Background())
			return err
		}},
		{name: "ospf", path: "/api/v4/vnet_ospf_commands", call: func(c *Client) error {
			_, err := c.VNetOSPFCommands.List(context.Background())
			return err
		}},
		{name: "eigrp", path: "/api/v4/vnet_eigrp_routers", call: func(c *Client) error {
			_, err := c.VNetEIGRPRouters.List(context.Background())
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"GET " + tc.path: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
			}))
			err := tc.call(client)
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
				t.Fatalf("expected API 500, got %v", err)
			}
		})
	}
}

func TestRouting_GetNotFound(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		resource string
		call     func(*Client) error
	}{
		{name: "bgp", path: "/api/v4/vnet_bgp/9", resource: "VNetBGP", call: func(c *Client) error {
			_, err := c.VNetBGP.Get(context.Background(), 9)
			return err
		}},
		{name: "router", path: "/api/v4/vnet_bgp_routers/9", resource: "VNetBGPRouter", call: func(c *Client) error {
			_, err := c.VNetBGPRouters.Get(context.Background(), 9)
			return err
		}},
		{name: "router command", path: "/api/v4/vnet_bgp_router_commands/9", resource: "VNetBGPRouterCommand", call: func(c *Client) error {
			_, err := c.VNetBGPRouterCommands.Get(context.Background(), 9)
			return err
		}},
		{name: "interface", path: "/api/v4/vnet_bgp_interfaces/9", resource: "VNetBGPInterface", call: func(c *Client) error {
			_, err := c.VNetBGPInterfaces.Get(context.Background(), 9)
			return err
		}},
		{name: "interface command", path: "/api/v4/vnet_bgp_interface_commands/9", resource: "VNetBGPInterfaceCommand", call: func(c *Client) error {
			_, err := c.VNetBGPInterfaceCommands.Get(context.Background(), 9)
			return err
		}},
		{name: "route map", path: "/api/v4/vnet_bgp_routemaps/9", resource: "VNetBGPRouteMap", call: func(c *Client) error {
			_, err := c.VNetBGPRouteMaps.Get(context.Background(), 9)
			return err
		}},
		{name: "route map command", path: "/api/v4/vnet_bgp_routemap_commands/9", resource: "VNetBGPRouteMapCommand", call: func(c *Client) error {
			_, err := c.VNetBGPRouteMapCommands.Get(context.Background(), 9)
			return err
		}},
		{name: "ip", path: "/api/v4/vnet_bgp_ip/9", resource: "VNetBGPIPCommand", call: func(c *Client) error {
			_, err := c.VNetBGPIPCommands.Get(context.Background(), 9)
			return err
		}},
		{name: "ospf", path: "/api/v4/vnet_ospf_commands/9", resource: "VNetOSPFCommand", call: func(c *Client) error {
			_, err := c.VNetOSPFCommands.Get(context.Background(), 9)
			return err
		}},
		{name: "eigrp", path: "/api/v4/vnet_eigrp_routers/9", resource: "VNetEIGRPRouter", call: func(c *Client) error {
			_, err := c.VNetEIGRPRouters.Get(context.Background(), 9)
			return err
		}},
		{name: "eigrp command", path: "/api/v4/vnet_eigrp_router_commands/9", resource: "VNetEIGRPRouterCommand", call: func(c *Client) error {
			_, err := c.VNetEIGRPRouterCommands.Get(context.Background(), 9)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var fields string
			client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
				"GET " + tc.path: func(w http.ResponseWriter, r *http.Request) {
					fields = r.URL.Query().Get("fields")
					w.WriteHeader(http.StatusNotFound)
				},
			}))
			err := tc.call(client)
			if !IsNotFoundError(err) {
				t.Fatalf("expected NotFoundError, got %v", err)
			}
			if got := err.Error(); !contains(got, tc.resource) {
				t.Fatalf("error %q does not name %s", got, tc.resource)
			}
			if fields == "" || fields == "most" {
				t.Fatalf("get fields = %q", fields)
			}
		})
	}
}

func contains(s, part string) bool {
	return len(s) >= len(part) && (s == part || len(part) == 0 || (func() bool {
		for i := 0; i+len(part) <= len(s); i++ {
			if s[i:i+len(part)] == part {
				return true
			}
		}
		return false
	})())
}

func TestRouting_CreateBodiesAndPendingRestart(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		key      int
		want     map[string]any
		absent   []string
		readBack any
		call     func(*Client) (int, *RoutingRestartStatus, error)
	}{
		{
			name: "bgp config",
			path: "/api/v4/vnet_bgp",
			key:  4,
			want: map[string]any{"vnet": float64(10)},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGP.Create(context.Background(), &VNetBGPCreateRequest{VNet: 10})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "bgp router",
			path: "/api/v4/vnet_bgp_routers",
			key:  8,
			want: map[string]any{"bgp": float64(4), "asn": float64(65000)},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPRouters.Create(context.Background(), &VNetBGPRouterCreateRequest{BGP: 4, ASN: 65000})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name:   "bgp router command omits bools",
			path:   "/api/v4/vnet_bgp_router_commands",
			key:    15,
			want:   map[string]any{"bgp_router": float64(8), "command": "neighbor", "params": "192.168.1.1 remote-as 65001"},
			absent: []string{"enabled", "no"},
			readBack: VNetBGPRouterCommand{
				Key: 15, BGPRouter: 8, Command: BGPRouterCommandNeighbor, Params: "192.168.1.1 remote-as 65001",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPRouterCommands.Create(context.Background(), &VNetBGPRouterCommandCreateRequest{
					BGPRouter: 8,
					Command:   BGPRouterCommandNeighbor,
					Params:    "192.168.1.1 remote-as 65001",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "bgp router command sends false bools",
			path: "/api/v4/vnet_bgp_router_commands",
			key:  15,
			want: map[string]any{"bgp_router": float64(8), "command": "network", "params": "", "enabled": false, "no": true},
			readBack: VNetBGPRouterCommand{
				Key: 15, BGPRouter: 8, Command: "network", Enabled: false, No: true,
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPRouterCommands.Create(context.Background(), &VNetBGPRouterCommandCreateRequest{
					BGPRouter: 8,
					Command:   BGPRouterCommandNetwork,
					Enabled:   boolPtr(false),
					No:        boolPtr(true),
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "bgp interface",
			path: "/api/v4/vnet_bgp_interfaces",
			key:  6,
			want: map[string]any{
				"bgp": float64(4), "name": "uplink", "ipaddress": "10.255.0.1",
				"network": "10.255.0.0/30", "interface_vnet": float64(20),
				"layer2_type": "vlan", "layer2_id": float64(100), "mtu": float64(9000),
				"description": "peer",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				layer2 := BGPInterfaceLayer2VLAN
				row, status, err := c.VNetBGPInterfaces.Create(context.Background(), &VNetBGPInterfaceCreateRequest{
					BGP: 4, Name: "uplink", IPAddress: "10.255.0.1", Network: "10.255.0.0/30",
					InterfaceVNet: 20, Layer2Type: &layer2, Layer2ID: intPtr(100), MTU: intPtr(9000),
					Description: "peer",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name:   "bgp interface omits optional fields",
			path:   "/api/v4/vnet_bgp_interfaces",
			key:    6,
			want:   map[string]any{"bgp": float64(4), "name": "uplink", "ipaddress": "10.255.0.1", "network": "10.255.0.0/30", "interface_vnet": float64(20)},
			absent: []string{"layer2_type", "layer2_id", "mtu", "description"},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPInterfaces.Create(context.Background(), &VNetBGPInterfaceCreateRequest{
					BGP: 4, Name: "uplink", IPAddress: "10.255.0.1", Network: "10.255.0.0/30", InterfaceVNet: 20,
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "interface command",
			path: "/api/v4/vnet_bgp_interface_commands",
			key:  16,
			want: map[string]any{"bgp_interface": float64(6), "command": "ospf", "params": "cost 10"},
			readBack: VNetBGPInterfaceCommand{
				Key: 16, BGPInterface: 6, Command: BGPInterfaceCommandOSPF, Params: "cost 10",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPInterfaceCommands.Create(context.Background(), &VNetBGPInterfaceCommandCreateRequest{
					BGPInterface: 6, Command: BGPInterfaceCommandOSPF, Params: "cost 10",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name:   "route map omits permit",
			path:   "/api/v4/vnet_bgp_routemaps",
			key:    3,
			want:   map[string]any{"bgp": float64(4), "tag": "IMPORT", "sequence": float64(10)},
			absent: []string{"permit"},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPRouteMaps.Create(context.Background(), &VNetBGPRouteMapCreateRequest{
					BGP: 4, Tag: "IMPORT", Sequence: 10,
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "route map deny",
			path: "/api/v4/vnet_bgp_routemaps",
			key:  3,
			want: map[string]any{"bgp": float64(4), "tag": "BLOCK", "sequence": float64(10), "permit": false},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPRouteMaps.Create(context.Background(), &VNetBGPRouteMapCreateRequest{
					BGP: 4, Tag: "BLOCK", Sequence: 10, Permit: boolPtr(false),
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "route map command",
			path: "/api/v4/vnet_bgp_routemap_commands",
			key:  17,
			want: map[string]any{"bgp_routemap": float64(3), "command": "match", "params": "as-path 1"},
			readBack: VNetBGPRouteMapCommand{
				Key: 17, BGPRouteMap: 3, Command: BGPRouteMapCommandMatch, Params: "as-path 1",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPRouteMapCommands.Create(context.Background(), &VNetBGPRouteMapCommandCreateRequest{
					BGPRouteMap: 3, Command: BGPRouteMapCommandMatch, Params: "as-path 1",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "ip command",
			path: "/api/v4/vnet_bgp_ip",
			key:  19,
			want: map[string]any{"bgp": float64(4), "command": "prefix-list", "params": "MY-PREFIX seq 10 permit 10.0.0.0/8 le 24"},
			readBack: VNetBGPIPCommand{
				Key: 19, BGP: 4, Command: BGPIPCommandPrefixList, Params: "MY-PREFIX seq 10 permit 10.0.0.0/8 le 24",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetBGPIPCommands.Create(context.Background(), &VNetBGPIPCommandCreateRequest{
					BGP: 4, Command: BGPIPCommandPrefixList, Params: "MY-PREFIX seq 10 permit 10.0.0.0/8 le 24",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "ospf command",
			path: "/api/v4/vnet_ospf_commands",
			key:  20,
			want: map[string]any{"bgp": float64(4), "command": "router-id", "params": "1.1.1.1"},
			readBack: VNetOSPFCommand{
				Key: 20, BGP: 4, Command: OSPFCommandRouterID, Params: "1.1.1.1",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetOSPFCommands.Create(context.Background(), &VNetOSPFCommandCreateRequest{
					BGP: 4, Command: OSPFCommandRouterID, Params: "1.1.1.1",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name: "eigrp router",
			path: "/api/v4/vnet_eigrp_routers",
			key:  2,
			want: map[string]any{"bgp": float64(4), "asn": float64(100)},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetEIGRPRouters.Create(context.Background(), &VNetEIGRPRouterCreateRequest{BGP: 4, ASN: 100})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
		{
			name:   "eigrp command omits bools",
			path:   "/api/v4/vnet_eigrp_router_commands",
			key:    18,
			want:   map[string]any{"eigrp_router": float64(2), "command": "network", "params": "10.0.0.0/24"},
			absent: []string{"enabled", "no"},
			readBack: VNetEIGRPRouterCommand{
				Key: 18, EIGRPRouter: 2, Command: EIGRPRouterCommandNetwork, Params: "10.0.0.0/24",
			},
			call: func(c *Client) (int, *RoutingRestartStatus, error) {
				row, status, err := c.VNetEIGRPRouterCommands.Create(context.Background(), &VNetEIGRPRouterCommandCreateRequest{
					EIGRPRouter: 2, Command: EIGRPRouterCommandNetwork, Params: "10.0.0.0/24",
				})
				if row == nil {
					return 0, status, err
				}
				return int(row.Key), status, err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			restarted := false
			routes := routingParentRoutes(true)
			routes["POST "+tc.path] = func(w http.ResponseWriter, r *http.Request) {
				assertJSONBody(t, decodeJSONBody(t, r), tc.want, tc.absent)
				jsonResponse(w, 200, map[string]any{"$key": tc.key})
			}
			if tc.readBack != nil {
				routes["GET "+tc.path+"/"+itoa(tc.key)] = func(w http.ResponseWriter, r *http.Request) {
					jsonResponse(w, 200, tc.readBack)
				}
			}
			routes["POST /api/v4/vnet_actions"] = func(w http.ResponseWriter, r *http.Request) {
				restarted = true
				w.WriteHeader(http.StatusOK)
			}
			client := newTestClient(t, apiMux(routes))
			key, status, err := tc.call(client)
			if err != nil {
				t.Fatal(err)
			}
			if key != tc.key {
				t.Fatalf("key = %d, want %d", key, tc.key)
			}
			assertPendingRestart(t, status)
			if restarted {
				t.Fatal("network was restarted without WithRestartNetwork")
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

func TestRouting_WithRestartNetwork(t *testing.T) {
	client := newTestClient(t, apiMux(mergeRoutes(routingParentRoutes(false), map[string]http.HandlerFunc{
		"POST /api/v4/vnet_bgp_routers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 8})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			body := decodeJSONBody(t, r)
			if body["action"] != "reset" || body["vnet"] != float64(10) {
				t.Fatalf("unexpected action: %#v", body)
			}
			params, _ := body["params"].(map[string]any)
			if params["apply"] != false {
				t.Fatalf("restart apply = %#v, want false", params["apply"])
			}
			w.WriteHeader(http.StatusOK)
		},
	})))

	row, status, err := client.VNetBGPRouters.Create(context.Background(), &VNetBGPRouterCreateRequest{
		BGP: 4, ASN: 65000,
	}, WithRestartNetwork())
	if err != nil {
		t.Fatal(err)
	}
	if row == nil || int(row.Key) != 8 {
		t.Fatalf("unexpected router: %+v", row)
	}
	if status == nil || !status.Restarted || status.Pending || status.NetworkID != 10 {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestRouting_RestartFailureReturnsRow(t *testing.T) {
	client := newTestClient(t, apiMux(mergeRoutes(routingParentRoutes(true), map[string]http.HandlerFunc{
		"POST /api/v4/vnet_ospf_commands": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"$key": 20})
		},
		"GET /api/v4/vnet_ospf_commands/20": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetOSPFCommand{Key: 20, BGP: 4, Command: "network", Params: "10.0.0.0/24 area 0"})
		},
		"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		},
	})))

	row, status, err := client.VNetOSPFCommands.Create(context.Background(), &VNetOSPFCommandCreateRequest{
		BGP: 4, Command: OSPFCommandNetwork, Params: "10.0.0.0/24 area 0",
	}, WithRestartNetwork())
	if err == nil {
		t.Fatal("expected restart error")
	}
	if row == nil || int(row.Key) != 20 {
		t.Fatalf("expected created command with the error, got %+v", row)
	}
	if status == nil || status.Restarted || !status.Pending || status.NetworkID != 10 {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestRouting_CreateServerError(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_bgp_ip": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	}))
	row, status, err := client.VNetBGPIPCommands.Create(context.Background(), &VNetBGPIPCommandCreateRequest{
		BGP: 4, Command: BGPIPCommandASPath, Params: "access-list 1 permit ^65001$",
	})
	if row != nil || status != nil || err == nil {
		t.Fatalf("row=%v status=%v err=%v", row, status, err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 500 {
		t.Fatalf("expected API 500, got %v", err)
	}
}

func TestRouting_CreateMissingKey(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"POST /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, map[string]any{"status": "ok"})
		},
	}))
	_, _, err := client.VNetBGP.Create(context.Background(), &VNetBGPCreateRequest{VNet: 10})
	if err == nil || !contains(err.Error(), "$key") {
		t.Fatalf("expected missing $key, got %v", err)
	}
}

func TestRouting_Validation(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{}))
	ctx := context.Background()
	cases := []struct {
		name string
		call func() error
	}{
		{name: "nil bgp", call: func() error {
			_, _, err := client.VNetBGP.Create(ctx, nil)
			return err
		}},
		{name: "vnet", call: func() error {
			_, _, err := client.VNetBGP.GetOrCreate(ctx, 0)
			return err
		}},
		{name: "nil router", call: func() error {
			_, _, err := client.VNetBGPRouters.Create(ctx, nil)
			return err
		}},
		{name: "router bgp", call: func() error {
			_, _, err := client.VNetBGPRouters.Create(ctx, &VNetBGPRouterCreateRequest{ASN: 1})
			return err
		}},
		{name: "router asn", call: func() error {
			_, _, err := client.VNetBGPRouters.Create(ctx, &VNetBGPRouterCreateRequest{BGP: 4, ASN: 0})
			return err
		}},
		{name: "empty router update", call: func() error {
			_, _, err := client.VNetBGPRouters.Update(ctx, 8, &VNetBGPRouterUpdateRequest{})
			return err
		}},
		{name: "nil command", call: func() error {
			_, _, err := client.VNetBGPRouterCommands.Create(ctx, &VNetBGPRouterCommandCreateRequest{BGPRouter: 8})
			return err
		}},
		{name: "interface name", call: func() error {
			_, _, err := client.VNetBGPInterfaces.Create(ctx, &VNetBGPInterfaceCreateRequest{BGP: 4, IPAddress: "10.0.0.1", Network: "10.0.0.0/30", InterfaceVNet: 2})
			return err
		}},
		{name: "interface mtu", call: func() error {
			_, _, err := client.VNetBGPInterfaces.Create(ctx, &VNetBGPInterfaceCreateRequest{
				BGP: 4, Name: "uplink", IPAddress: "10.0.0.1", Network: "10.0.0.0/30", InterfaceVNet: 2, MTU: intPtr(999),
			})
			return err
		}},
		{name: "interface layer2", call: func() error {
			bad := "trunk"
			_, _, err := client.VNetBGPInterfaces.Create(ctx, &VNetBGPInterfaceCreateRequest{
				BGP: 4, Name: "uplink", IPAddress: "10.0.0.1", Network: "10.0.0.0/30", InterfaceVNet: 2, Layer2Type: &bad,
			})
			return err
		}},
		{name: "route map sequence", call: func() error {
			_, _, err := client.VNetBGPRouteMaps.Create(ctx, &VNetBGPRouteMapCreateRequest{BGP: 4, Tag: "IMPORT", Sequence: 0})
			return err
		}},
		{name: "eigrp asn", call: func() error {
			_, _, err := client.VNetEIGRPRouters.Create(ctx, &VNetEIGRPRouterCreateRequest{BGP: 4, ASN: 70000})
			return err
		}},
		{name: "empty ip update", call: func() error {
			_, _, err := client.VNetBGPIPCommands.Update(ctx, 1, nil)
			return err
		}},
		{name: "list router parent", call: func() error {
			_, err := client.VNetBGPRouters.ListByBGP(ctx, 0)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !IsValidationError(err) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

func TestValidateBGPASNBounds(t *testing.T) {
	if err := validateBGPASN(1); err != nil {
		t.Fatal(err)
	}
	if err := validateBGPASN(0); !IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
	// On 64-bit platforms int can represent an ASN above the 32-bit protocol maximum.
	if maxInt := int64(int(^uint(0) >> 1)); maxInt > maxBGPASN {
		if err := validateBGPASN(int(maxBGPASN + 1)); !IsValidationError(err) {
			t.Fatalf("expected upper bound error, got %v", err)
		}
	}
}

func TestVNetBGP_GetOrCreate(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		posted := false
		client := newTestClient(t, apiMux(mergeRoutes(routingParentRoutes(true), map[string]http.HandlerFunc{
			"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("filter") != "vnet eq 10" {
					t.Errorf("filter = %s", r.URL.Query().Get("filter"))
				}
				jsonResponse(w, 200, []VNetBGP{{Key: 4, VNet: 10}})
			},
			"POST /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				posted = true
				w.WriteHeader(http.StatusOK)
			},
			"POST /api/v4/vnet_actions": func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("existing config was restarted")
			},
		})))
		row, status, err := client.VNetBGP.GetOrCreate(context.Background(), 10, WithRestartNetwork())
		if err != nil {
			t.Fatal(err)
		}
		if row == nil || int(row.Key) != 4 || int(row.VNet) != 10 || status != nil || posted {
			t.Fatalf("row=%+v status=%+v posted=%v", row, status, posted)
		}
	})

	t.Run("create", func(t *testing.T) {
		client := newTestClient(t, apiMux(mergeRoutes(routingParentRoutes(true), map[string]http.HandlerFunc{
			"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []VNetBGP{})
			},
			"POST /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				assertJSONBody(t, decodeJSONBody(t, r), map[string]any{"vnet": float64(10)}, nil)
				jsonResponse(w, 200, map[string]any{"$key": 4})
			},
		})))
		row, status, err := client.VNetBGP.GetOrCreate(context.Background(), 10)
		if err != nil {
			t.Fatal(err)
		}
		if row == nil || int(row.Key) != 4 {
			t.Fatalf("unexpected row: %+v", row)
		}
		assertPendingRestart(t, status)
	})

	t.Run("ambiguous", func(t *testing.T) {
		client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
			"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []VNetBGP{{Key: 4, VNet: 10}, {Key: 5, VNet: 10}})
			},
		}))
		_, _, err := client.VNetBGP.GetOrCreate(context.Background(), 10)
		if !IsAmbiguousNameError(err) {
			t.Fatalf("expected AmbiguousNameError, got %v", err)
		}
	})
}

func TestRouting_ListByNetwork(t *testing.T) {
	t.Run("missing config", func(t *testing.T) {
		client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
			"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []VNetBGP{})
			},
			"GET /api/v4/vnet_bgp_routers": func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("routers were listed without a bgp config")
			},
			"GET /api/v4/vnet_ospf_commands": func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("ospf commands were listed without a bgp config")
			},
		}))
		routers, err := client.VNetBGPRouters.ListByNetwork(context.Background(), 10)
		if err != nil || len(routers) != 0 {
			t.Fatalf("routers=%v err=%v", routers, err)
		}
		cmds, err := client.VNetOSPFCommands.ListByNetwork(context.Background(), 10)
		if err != nil || len(cmds) != 0 {
			t.Fatalf("commands=%v err=%v", cmds, err)
		}
	})

	t.Run("found", func(t *testing.T) {
		client := newTestClient(t, apiMux(mergeRoutes(routingParentRoutes(true), map[string]http.HandlerFunc{
			"GET /api/v4/vnet_bgp": func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, []VNetBGP{{Key: 4, VNet: 10}})
			},
			"GET /api/v4/vnet_eigrp_routers": func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("filter") != "bgp eq 4" {
					t.Errorf("filter = %s", r.URL.Query().Get("filter"))
				}
				jsonResponse(w, 200, []VNetEIGRPRouter{{Key: 2, BGP: 4, ASN: 100}})
			},
		})))
		rows, err := client.VNetEIGRPRouters.ListByNetwork(context.Background(), 10)
		if err != nil || len(rows) != 1 || rows[0].ASN != 100 {
			t.Fatalf("rows=%+v err=%v", rows, err)
		}
	})
}

func TestRouting_Lookups(t *testing.T) {
	client := newTestClient(t, apiMux(mergeRoutes(routingParentRoutes(false), map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bgp_routers": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("filter") != "bgp eq 4 and asn eq 65000" {
				t.Errorf("router filter = %s", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []VNetBGPRouter{{Key: 8, BGP: 4, ASN: 65000}})
		},
		"GET /api/v4/vnet_bgp_interfaces": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("filter") != `bgp eq 4 and name eq 'peer\'s'` {
				t.Errorf("interface filter = %s", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []VNetBGPInterface{{Key: 6, BGP: 4, Name: "peer's"}})
		},
		"GET /api/v4/vnet_bgp_interfaces/6": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, VNetBGPInterface{Key: 6, BGP: 4, Name: "peer's"})
		},
		"GET /api/v4/vnet_bgp_routemaps": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("filter") != "bgp eq 4 and tag eq 'IMPORT' and sequence eq 10" {
				t.Errorf("routemap filter = %s", r.URL.Query().Get("filter"))
			}
			jsonResponse(w, 200, []VNetBGPRouteMap{{Key: 3, BGP: 4, Tag: "IMPORT", Sequence: 10}})
		},
		"GET /api/v4/vnet_eigrp_routers": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, 200, []VNetEIGRPRouter{{Key: 2, BGP: 4, ASN: 100}, {Key: 7, BGP: 4, ASN: 100}})
		},
	})))

	router, err := client.VNetBGPRouters.GetByASN(context.Background(), 4, 65000)
	if err != nil || int(router.Key) != 8 {
		t.Fatalf("router=%+v err=%v", router, err)
	}
	iface, err := client.VNetBGPInterfaces.GetByName(context.Background(), 4, "peer's")
	if err != nil || iface.Name != "peer's" {
		t.Fatalf("interface=%+v err=%v", iface, err)
	}
	routeMap, err := client.VNetBGPRouteMaps.GetByTagAndSequence(context.Background(), 4, "IMPORT", 10)
	if err != nil || int(routeMap.Key) != 3 {
		t.Fatalf("routemap=%+v err=%v", routeMap, err)
	}
	_, err = client.VNetEIGRPRouters.GetByASN(context.Background(), 4, 100)
	if !IsAmbiguousNameError(err) {
		t.Fatalf("expected AmbiguousNameError, got %v", err)
	}
}

func TestRouting_Updates(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		want   map[string]any
		absent []string
		call   func(*Client) (*RoutingRestartStatus, error)
	}{
		{
			name: "router asn",
			path: "/api/v4/vnet_bgp_routers/8",
			want: map[string]any{"asn": float64(65001)},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetBGPRouters.Update(context.Background(), 8, &VNetBGPRouterUpdateRequest{ASN: intPtr(65001)})
				return status, err
			},
		},
		{
			name:   "router command",
			path:   "/api/v4/vnet_bgp_router_commands/15",
			want:   map[string]any{"command": "network", "no": true},
			absent: []string{"enabled", "params"},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetBGPRouterCommands.Update(context.Background(), 15, &VNetBGPRouterCommandUpdateRequest{
					Command: strPtr(BGPRouterCommandNetwork),
					No:      boolPtr(true),
				})
				return status, err
			},
		},
		{
			name: "interface",
			path: "/api/v4/vnet_bgp_interfaces/6",
			want: map[string]any{"description": "", "mtu": float64(1500)},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				empty := ""
				_, status, err := c.VNetBGPInterfaces.Update(context.Background(), 6, &VNetBGPInterfaceUpdateRequest{
					Description: &empty,
					MTU:         intPtr(1500),
				})
				return status, err
			},
		},
		{
			name: "route map",
			path: "/api/v4/vnet_bgp_routemaps/3",
			want: map[string]any{"permit": false, "sequence": float64(20)},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetBGPRouteMaps.Update(context.Background(), 3, &VNetBGPRouteMapUpdateRequest{
					Permit:   boolPtr(false),
					Sequence: intPtr(20),
				})
				return status, err
			},
		},
		{
			name: "ip",
			path: "/api/v4/vnet_bgp_ip/19",
			want: map[string]any{"params": "access-list 1 permit ^65001$"},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetBGPIPCommands.Update(context.Background(), 19, &VNetBGPIPCommandUpdateRequest{
					Params: strPtr("access-list 1 permit ^65001$"),
				})
				return status, err
			},
		},
		{
			name: "ospf",
			path: "/api/v4/vnet_ospf_commands/20",
			want: map[string]any{"command": "area"},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetOSPFCommands.Update(context.Background(), 20, &VNetOSPFCommandUpdateRequest{
					Command: strPtr(OSPFCommandArea),
				})
				return status, err
			},
		},
		{
			name: "eigrp",
			path: "/api/v4/vnet_eigrp_routers/2",
			want: map[string]any{"asn": float64(200)},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetEIGRPRouters.Update(context.Background(), 2, &VNetEIGRPRouterUpdateRequest{ASN: intPtr(200)})
				return status, err
			},
		},
		{
			name: "eigrp command",
			path: "/api/v4/vnet_eigrp_router_commands/18",
			want: map[string]any{"enabled": false},
			call: func(c *Client) (*RoutingRestartStatus, error) {
				_, status, err := c.VNetEIGRPRouterCommands.Update(context.Background(), 18, &VNetEIGRPRouterCommandUpdateRequest{
					Enabled: boolPtr(false),
				})
				return status, err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			routes := routingParentRoutes(true)
			routes["PUT "+tc.path] = func(w http.ResponseWriter, r *http.Request) {
				assertJSONBody(t, decodeJSONBody(t, r), tc.want, tc.absent)
				w.WriteHeader(http.StatusOK)
			}
			routes["GET /api/v4/vnet_bgp_router_commands/15"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetBGPRouterCommand{Key: 15, BGPRouter: 8, Command: "network", No: true})
			}
			routes["GET /api/v4/vnet_bgp_ip/19"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetBGPIPCommand{Key: 19, BGP: 4, Command: "as-path"})
			}
			routes["GET /api/v4/vnet_ospf_commands/20"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetOSPFCommand{Key: 20, BGP: 4, Command: "area"})
			}
			routes["GET /api/v4/vnet_eigrp_router_commands/18"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetEIGRPRouterCommand{Key: 18, EIGRPRouter: 2, Enabled: false})
			}
			client := newTestClient(t, apiMux(routes))
			status, err := tc.call(client)
			if err != nil {
				t.Fatal(err)
			}
			assertPendingRestart(t, status)
		})
	}
}

func TestRouting_UpdateNotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"PUT /api/v4/vnet_bgp_interfaces/9": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}))
	_, _, err := client.VNetBGPInterfaces.Update(context.Background(), 9, &VNetBGPInterfaceUpdateRequest{Name: strPtr("uplink")})
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestRouting_Deletes(t *testing.T) {
	cases := []struct {
		name string
		path string
		call func(*Client) (*RoutingRestartStatus, error)
	}{
		{name: "bgp", path: "/api/v4/vnet_bgp/4", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGP.Delete(context.Background(), 4)
		}},
		{name: "router", path: "/api/v4/vnet_bgp_routers/8", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPRouters.Delete(context.Background(), 8)
		}},
		{name: "interface", path: "/api/v4/vnet_bgp_interfaces/6", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPInterfaces.Delete(context.Background(), 6)
		}},
		{name: "route map", path: "/api/v4/vnet_bgp_routemaps/3", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPRouteMaps.Delete(context.Background(), 3)
		}},
		{name: "ip", path: "/api/v4/vnet_bgp_ip/19", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPIPCommands.Delete(context.Background(), 19)
		}},
		{name: "ospf", path: "/api/v4/vnet_ospf_commands/20", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetOSPFCommands.Delete(context.Background(), 20)
		}},
		{name: "eigrp", path: "/api/v4/vnet_eigrp_routers/2", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetEIGRPRouters.Delete(context.Background(), 2)
		}},
		{name: "router command", path: "/api/v4/vnet_bgp_router_commands/15", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPRouterCommands.Delete(context.Background(), 15)
		}},
		{name: "interface command", path: "/api/v4/vnet_bgp_interface_commands/16", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPInterfaceCommands.Delete(context.Background(), 16)
		}},
		{name: "route map command", path: "/api/v4/vnet_bgp_routemap_commands/17", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetBGPRouteMapCommands.Delete(context.Background(), 17)
		}},
		{name: "eigrp command", path: "/api/v4/vnet_eigrp_router_commands/18", call: func(c *Client) (*RoutingRestartStatus, error) {
			return c.VNetEIGRPRouterCommands.Delete(context.Background(), 18)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deleted := false
			routes := routingParentRoutes(true)
			routes["DELETE "+tc.path] = func(w http.ResponseWriter, r *http.Request) {
				deleted = true
				w.WriteHeader(http.StatusOK)
			}
			routes["GET /api/v4/vnet_bgp_ip/19"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetBGPIPCommand{Key: 19, BGP: 4})
			}
			routes["GET /api/v4/vnet_ospf_commands/20"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetOSPFCommand{Key: 20, BGP: 4})
			}
			routes["GET /api/v4/vnet_bgp_router_commands/15"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetBGPRouterCommand{Key: 15, BGPRouter: 8})
			}
			routes["GET /api/v4/vnet_bgp_interface_commands/16"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetBGPInterfaceCommand{Key: 16, BGPInterface: 6})
			}
			routes["GET /api/v4/vnet_bgp_routemap_commands/17"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetBGPRouteMapCommand{Key: 17, BGPRouteMap: 3})
			}
			routes["GET /api/v4/vnet_eigrp_router_commands/18"] = func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, 200, VNetEIGRPRouterCommand{Key: 18, EIGRPRouter: 2})
			}
			client := newTestClient(t, apiMux(routes))
			status, err := tc.call(client)
			if err != nil {
				t.Fatal(err)
			}
			if !deleted {
				t.Fatal("delete was not sent")
			}
			assertPendingRestart(t, status)
		})
	}
}

func TestRouting_DeleteNotFound(t *testing.T) {
	client := newTestClient(t, apiMux(map[string]http.HandlerFunc{
		"GET /api/v4/vnet_bgp_routemap_commands/9": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}))
	_, err := client.VNetBGPRouteMapCommands.Delete(context.Background(), 9)
	if !IsNotFoundError(err) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

func TestVNetBGPInterface_NullOptionals(t *testing.T) {
	var iface VNetBGPInterface
	raw := []byte(`{"$key":"6","bgp":4,"name":"uplink","layer2_id":null,"interface_vnet":null,"bgp_vnet":null,"nic":null}`)
	if err := json.Unmarshal(raw, &iface); err != nil {
		t.Fatal(err)
	}
	if int(iface.Key) != 6 || int(iface.BGP) != 4 || iface.Name != "uplink" {
		t.Fatalf("unexpected interface: %+v", iface)
	}
	if iface.Layer2ID != nil || iface.InterfaceVNet != nil || iface.BGPVNet != nil || iface.NIC != nil {
		t.Fatalf("expected nil optionals, got %+v", iface)
	}
}

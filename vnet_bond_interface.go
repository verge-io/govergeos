package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// VNetBondInterfaceService reads and updates bond members (vnet_bond_interfaces).
// The primary flag is requested only when the server is VergeOS 26.2 or later.
type VNetBondInterfaceService struct {
	client *Client
}

func (s *VNetBondInterfaceService) listFields() string {
	return appendVersionFields(vnetBondInterfaceListFields, vnetBondInterface262Fields, s.client.atLeast262())
}

// List returns bond members.
func (s *VNetBondInterfaceService) List(ctx context.Context, opts ...ListOption) ([]VNetBondInterface, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = s.listFields()
	}
	var rows []VNetBondInterface
	if err := s.client.get(ctx, "/vnet_bond_interfaces", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByNetwork returns the bond members of one network.
func (s *VNetBondInterfaceService) ListByNetwork(ctx context.Context, networkID int, opts ...ListOption) ([]VNetBondInterface, error) {
	opts = append(opts, WithFilter(fmt.Sprintf("vnet eq %d", networkID)))
	return s.List(ctx, opts...)
}

// Get returns one bond member.
func (s *VNetBondInterfaceService) Get(ctx context.Context, id int) (*VNetBondInterface, error) {
	params := url.Values{}
	params.Set("fields", s.listFields())

	var row VNetBondInterface
	endpoint := fmt.Sprintf("/vnet_bond_interfaces/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "VNetBondInterface", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

// Update sets fields on a bond member. Primary is the 26.2 flag.
func (s *VNetBondInterfaceService) Update(ctx context.Context, id int, req *VNetBondInterfaceUpdateRequest) (*VNetBondInterface, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	endpoint := fmt.Sprintf("/vnet_bond_interfaces/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

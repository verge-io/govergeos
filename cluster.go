package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// ClusterService handles cluster read operations.
type ClusterService struct {
	client *Client
}

// List returns all clusters, with optional filtering and pagination.
func (s *ClusterService) List(ctx context.Context, opts ...ListOption) ([]Cluster, error) {
	options := applyListOptions(opts)

	if options.Fields == "most" {
		options.Fields = clusterListFields
	}

	params := options.toQueryParams()

	var clusters []Cluster
	if err := s.client.get(ctx, "/clusters", params, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}

// Get returns a single cluster by ID.
func (s *ClusterService) Get(ctx context.Context, id int) (*Cluster, error) {
	params := url.Values{}
	params.Set("fields", clusterListFields)

	var cluster Cluster
	endpoint := fmt.Sprintf("/clusters/%d", id)
	if err := s.client.get(ctx, endpoint, params, &cluster); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Cluster", ID: id}
		}
		return nil, err
	}

	return &cluster, nil
}

// GetByName returns a cluster by name.
func (s *ClusterService) GetByName(ctx context.Context, name string) (*Cluster, error) {
	clusters, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", name)))
	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, &NotFoundError{Resource: "Cluster", ID: name}
	}

	return &clusters[0], nil
}

// GetStatus returns detailed status for a cluster.
func (s *ClusterService) GetStatus(ctx context.Context, id int) (*ClusterStatus, error) {
	params := url.Values{}
	params.Set("fields", "status")

	var cluster struct {
		Status ClusterStatus `json:"status"`
	}
	endpoint := fmt.Sprintf("/clusters/%d", id)
	if err := s.client.get(ctx, endpoint, params, &cluster); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "Cluster", ID: id}
		}
		return nil, err
	}

	return &cluster.Status, nil
}

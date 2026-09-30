package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// BillingService reads usage records (billing) and asks VergeOS to write one.
type BillingService struct {
	client *Client
}

// List returns billing records, newest first.
//
// A caller sort replaces the default "-created" order.
func (s *BillingService) List(ctx context.Context, opts ...ListOption) ([]BillingRecord, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = billingListFields
	}
	if options.Sort == "" {
		options.Sort = "-created"
	}

	var records []BillingRecord
	if err := s.client.get(ctx, "/billing", options.toQueryParams(), &records); err != nil {
		return nil, err
	}
	return records, nil
}

// ListCreated returns records whose created timestamp falls in [since, until].
//
// A zero bound is omitted. Results are newest first unless a sort is set.
func (s *BillingService) ListCreated(ctx context.Context, since, until int64, opts ...ListOption) ([]BillingRecord, error) {
	if since != 0 {
		opts = append(opts, WithFilter(fmt.Sprintf("created ge %d", since)))
	}
	if until != 0 {
		opts = append(opts, WithFilter(fmt.Sprintf("created le %d", until)))
	}
	return s.List(ctx, opts...)
}

// Get returns one billing record by key.
func (s *BillingService) Get(ctx context.Context, id int) (*BillingRecord, error) {
	params := url.Values{}
	params.Set("fields", billingListFields)

	var record BillingRecord
	endpoint := fmt.Sprintf("/billing/%d", id)
	if err := s.client.get(ctx, endpoint, params, &record); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "BillingRecord", ID: id}
		}
		return nil, err
	}
	return &record, nil
}

// GetLatest returns the newest billing record.
func (s *BillingService) GetLatest(ctx context.Context, opts ...ListOption) (*BillingRecord, error) {
	opts = append(opts, WithLimit(1), WithSort("-created"))
	records, err := s.List(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, &NotFoundError{Resource: "BillingRecord", ID: "latest"}
	}
	return &records[0], nil
}

// Generate asks VergeOS to write a billing record for current usage.
//
// The action is POST /billing_actions with action "generate".
func (s *BillingService) Generate(ctx context.Context) error {
	body := struct {
		Action string `json:"action"`
	}{Action: "generate"}
	if err := s.client.post(ctx, "/billing_actions", body, nil); err != nil {
		return fmt.Errorf("vergeos: failed to generate billing record: %w", err)
	}
	return nil
}

// GetSummary averages the billing records in [since, until].
//
// A zero bound is omitted. An empty range returns a zero summary.
// GPU totals come from the newest record.
func (s *BillingService) GetSummary(ctx context.Context, since, until int64) (*BillingSummary, error) {
	records, err := s.ListCreated(ctx, since, until)
	if err != nil {
		return nil, err
	}
	summary := &BillingSummary{RecordCount: len(records)}
	if len(records) == 0 {
		return summary, nil
	}

	var cpu, ram, storage, gpus, vgpus float64
	for _, record := range records {
		cpu += record.CPUUtilization()
		ram += record.RAMUtilization()
		used := record.TotalStorageUsedGB()
		storage += used
		gpus += float64(record.GPUs)
		vgpus += float64(record.VGPUs)
		if record.UsedCores > summary.PeakCPUCores {
			summary.PeakCPUCores = record.UsedCores
		}
		if gb := record.UsedRAMGB(); gb > summary.PeakRAMGB {
			summary.PeakRAMGB = gb
		}
		if used > summary.PeakStorageUsedGB {
			summary.PeakStorageUsedGB = used
		}
	}
	n := float64(len(records))
	summary.AvgCPUUtilization = cpu / n
	summary.AvgRAMUtilization = ram / n
	summary.AvgStorageUsedGB = storage / n
	summary.TotalGPUs = records[0].GPUsTotal
	summary.AvgGPUsUsed = gpus / n
	summary.TotalVGPUs = records[0].VGPUsTotal
	summary.AvgVGPUsUsed = vgpus / n
	return summary, nil
}

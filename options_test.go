package vergeos

import (
	"testing"
)

func TestApplyListOptions_Defaults(t *testing.T) {
	opts := applyListOptions(nil)
	if opts.Fields != "most" {
		t.Errorf("expected default fields 'most', got %q", opts.Fields)
	}
	if opts.Filter != "" {
		t.Errorf("expected empty filter, got %q", opts.Filter)
	}
	if opts.Sort != "" {
		t.Errorf("expected empty sort, got %q", opts.Sort)
	}
	if opts.Limit != 0 {
		t.Errorf("expected limit 0, got %d", opts.Limit)
	}
	if opts.Offset != 0 {
		t.Errorf("expected offset 0, got %d", opts.Offset)
	}
}

func TestApplyListOptions_AllOptions(t *testing.T) {
	opts := applyListOptions([]ListOption{
		WithFilter("name eq 'test'"),
		WithFields("all"),
		WithSort("-created"),
		WithLimit(50),
		WithOffset(10),
	})
	if opts.Filter != "name eq 'test'" {
		t.Errorf("expected filter \"name eq 'test'\", got %q", opts.Filter)
	}
	if opts.Fields != "all" {
		t.Errorf("expected fields 'all', got %q", opts.Fields)
	}
	if opts.Sort != "-created" {
		t.Errorf("expected sort '-created', got %q", opts.Sort)
	}
	if opts.Limit != 50 {
		t.Errorf("expected limit 50, got %d", opts.Limit)
	}
	if opts.Offset != 10 {
		t.Errorf("expected offset 10, got %d", opts.Offset)
	}
}

func TestApplyListOptions_FilterCombination(t *testing.T) {
	opts := applyListOptions([]ListOption{
		WithFilter("first"),
		WithFilter("second"),
	})
	if opts.Filter != "first and second" {
		t.Errorf("expected combined filter, got %q", opts.Filter)
	}
}

func TestApplyListOptions_TripleFilterCombination(t *testing.T) {
	opts := applyListOptions([]ListOption{
		WithFilter("a eq 1"),
		WithFilter("b eq 2"),
		WithFilter("c eq 3"),
	})
	expected := "a eq 1 and b eq 2 and c eq 3"
	if opts.Filter != expected {
		t.Errorf("expected %q, got %q", expected, opts.Filter)
	}
}

func TestToQueryParams_Empty(t *testing.T) {
	opts := &ListOptions{}
	params := opts.toQueryParams()
	if len(params) != 0 {
		t.Errorf("expected no params for empty options, got %v", params)
	}
}

func TestToQueryParams_FieldsOnly(t *testing.T) {
	opts := &ListOptions{Fields: "name,enabled"}
	params := opts.toQueryParams()
	if params.Get("fields") != "name,enabled" {
		t.Errorf("expected fields 'name,enabled', got %q", params.Get("fields"))
	}
	if params.Get("filter") != "" {
		t.Error("expected no filter param")
	}
}

func TestToQueryParams_AllParams(t *testing.T) {
	opts := &ListOptions{
		Fields: "all",
		Filter: "enabled eq true",
		Sort:   "-name",
		Limit:  25,
		Offset: 5,
	}
	params := opts.toQueryParams()

	if params.Get("fields") != "all" {
		t.Errorf("fields: got %q", params.Get("fields"))
	}
	if params.Get("filter") != "enabled eq true" {
		t.Errorf("filter: got %q", params.Get("filter"))
	}
	if params.Get("sort") != "-name" {
		t.Errorf("sort: got %q", params.Get("sort"))
	}
	if params.Get("limit") != "25" {
		t.Errorf("limit: got %q", params.Get("limit"))
	}
	if params.Get("offset") != "5" {
		t.Errorf("offset: got %q", params.Get("offset"))
	}
}

func TestToQueryParams_ZeroLimitAndOffset(t *testing.T) {
	opts := &ListOptions{
		Fields: "most",
		Limit:  0,
		Offset: 0,
	}
	params := opts.toQueryParams()

	if params.Get("limit") != "" {
		t.Error("zero limit should not produce a param")
	}
	if params.Get("offset") != "" {
		t.Error("zero offset should not produce a param")
	}
}

func TestWithFilter(t *testing.T) {
	opts := &ListOptions{}
	WithFilter("name eq 'foo'")(opts)
	if opts.Filter != "name eq 'foo'" {
		t.Errorf("got %q", opts.Filter)
	}
}

func TestWithFields(t *testing.T) {
	opts := &ListOptions{}
	WithFields("name,ram")(opts)
	if opts.Fields != "name,ram" {
		t.Errorf("got %q", opts.Fields)
	}
}

func TestWithSort(t *testing.T) {
	opts := &ListOptions{}
	WithSort("-created")(opts)
	if opts.Sort != "-created" {
		t.Errorf("got %q", opts.Sort)
	}
}

func TestWithLimit(t *testing.T) {
	opts := &ListOptions{}
	WithLimit(100)(opts)
	if opts.Limit != 100 {
		t.Errorf("got %d", opts.Limit)
	}
}

func TestWithOffset(t *testing.T) {
	opts := &ListOptions{}
	WithOffset(20)(opts)
	if opts.Offset != 20 {
		t.Errorf("got %d", opts.Offset)
	}
}

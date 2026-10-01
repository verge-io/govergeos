package vergeos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// TaskScheduleService handles task_schedules.
type TaskScheduleService struct {
	client *Client
}

// List returns task schedules.
func (s *TaskScheduleService) List(ctx context.Context, opts ...ListOption) ([]TaskSchedule, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = taskScheduleListFields
	}
	var rows []TaskSchedule
	if err := s.client.get(ctx, "/task_schedules", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListEnabled returns schedules with enabled eq true.
func (s *TaskScheduleService) ListEnabled(ctx context.Context, opts ...ListOption) ([]TaskSchedule, error) {
	filterOpts := []ListOption{WithFilter("enabled eq true")}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// ListDisabled returns schedules with enabled eq false.
func (s *TaskScheduleService) ListDisabled(ctx context.Context, opts ...ListOption) ([]TaskSchedule, error) {
	filterOpts := []ListOption{WithFilter("enabled eq false")}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns one schedule by row key.
func (s *TaskScheduleService) Get(ctx context.Context, id int) (*TaskSchedule, error) {
	params := url.Values{}
	params.Set("fields", taskScheduleListFields)

	var row TaskSchedule
	endpoint := fmt.Sprintf("/task_schedules/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TaskSchedule", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

// GetByName returns the schedule with this name.
//
// More than one row with the name is AmbiguousNameError.
func (s *TaskScheduleService) GetByName(ctx context.Context, name string) (*TaskSchedule, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "TaskSchedule", ID: name}
	}
	if err := requireUniqueName("TaskSchedule", name, rows, func(row TaskSchedule) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("TaskSchedule", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(rows[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("TaskSchedule", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a schedule and reads it back.
func (s *TaskScheduleService) Create(ctx context.Context, req *TaskScheduleCreateRequest) (*TaskSchedule, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if err := validateTimeOfDay("start_time_of_day", req.StartTimeOfDay); err != nil {
		return nil, err
	}
	if err := validateTimeOfDay("end_time_of_day", req.EndTimeOfDay); err != nil {
		return nil, err
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/task_schedules", req.wire(), &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update updates a schedule and reads it back.
func (s *TaskScheduleService) Update(ctx context.Context, id int, req *TaskScheduleUpdateRequest) (*TaskSchedule, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	if err := validateTimeOfDay("start_time_of_day", req.StartTimeOfDay); err != nil {
		return nil, err
	}
	if err := validateTimeOfDay("end_time_of_day", req.EndTimeOfDay); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/task_schedules/%d", id)
	if err := s.client.put(ctx, endpoint, req, nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete deletes a schedule.
//
// A schedule that still has triggers cannot be deleted; the platform returns
// that as a conflict.
func (s *TaskScheduleService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/task_schedules/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "TaskSchedule", ID: id}
		}
		return err
	}
	return nil
}

// Enable sets enabled to true and returns the schedule.
func (s *TaskScheduleService) Enable(ctx context.Context, id int) (*TaskSchedule, error) {
	enabled := true
	return s.Update(ctx, id, &TaskScheduleUpdateRequest{Enabled: &enabled})
}

// Disable sets enabled to false and returns the schedule.
func (s *TaskScheduleService) Disable(ctx context.Context, id int) (*TaskSchedule, error) {
	enabled := false
	return s.Update(ctx, id, &TaskScheduleUpdateRequest{Enabled: &enabled})
}

// GetSchedule returns upcoming run times for a schedule.
//
// The call is PUT /task_schedules/{id}?action=get_schedule. query nil asks
// for 100 times. The platform returns those times under "times" or
// "schedule", or as a JSON array. An empty body is an empty list.
func (s *TaskScheduleService) GetSchedule(ctx context.Context, id int, query *TaskScheduleQuery) ([]map[string]any, error) {
	body := taskScheduleQueryBody{Max: taskScheduleQueryDefaultMax}
	if query != nil {
		body.Max = clampScheduleMax(query.MaxResults)
		body.StartTime = query.StartTime
		body.EndTime = query.EndTime
	}
	endpoint := fmt.Sprintf("/task_schedules/%d?action=get_schedule", id)
	raw, err := s.client.putRaw(ctx, endpoint, body)
	if err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TaskSchedule", ID: id}
		}
		return nil, err
	}
	return scheduleTimes(raw)
}

// taskScheduleCreateBody is the JSON sent for a create. Defaults are filled
// so a partial request matches pyVergeOS, which always posts them.
type taskScheduleCreateBody struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Enabled         bool   `json:"enabled"`
	RepeatEvery     string `json:"repeat_every"`
	RepeatIteration int    `json:"repeat_iteration"`
	StartDate       string `json:"start_date,omitempty"`
	EndDate         string `json:"end_date,omitempty"`
	StartTimeOfDay  int    `json:"start_time_of_day"`
	EndTimeOfDay    int    `json:"end_time_of_day"`
	DayOfMonth      string `json:"day_of_month"`
	Monday          bool   `json:"monday"`
	Tuesday         bool   `json:"tuesday"`
	Wednesday       bool   `json:"wednesday"`
	Thursday        bool   `json:"thursday"`
	Friday          bool   `json:"friday"`
	Saturday        bool   `json:"saturday"`
	Sunday          bool   `json:"sunday"`
	Task            *int   `json:"task,omitempty"`
}

type taskScheduleQueryBody struct {
	Max       int    `json:"max"`
	StartTime *int64 `json:"start_time,omitempty"`
	EndTime   *int64 `json:"end_time,omitempty"`
}

func (req *TaskScheduleCreateRequest) wire() taskScheduleCreateBody {
	repeat := req.RepeatEvery
	if repeat == "" {
		repeat = TaskScheduleRepeatHour
	}
	iteration := 1
	if req.RepeatIteration != nil {
		iteration = *req.RepeatIteration
	}
	start := 0
	if req.StartTimeOfDay != nil {
		start = *req.StartTimeOfDay
	}
	end := TaskScheduleEndOfDay
	if req.EndTimeOfDay != nil {
		end = *req.EndTimeOfDay
	}
	day := req.DayOfMonth
	if day == "" {
		day = TaskScheduleDayOfMonthStartDate
	}
	return taskScheduleCreateBody{
		Name:            req.Name,
		Description:     req.Description,
		Enabled:         boolDefault(req.Enabled, true),
		RepeatEvery:     repeat,
		RepeatIteration: iteration,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		StartTimeOfDay:  start,
		EndTimeOfDay:    end,
		DayOfMonth:      day,
		Monday:          boolDefault(req.Monday, true),
		Tuesday:         boolDefault(req.Tuesday, true),
		Wednesday:       boolDefault(req.Wednesday, true),
		Thursday:        boolDefault(req.Thursday, true),
		Friday:          boolDefault(req.Friday, true),
		Saturday:        boolDefault(req.Saturday, true),
		Sunday:          boolDefault(req.Sunday, true),
		Task:            req.Task,
	}
}

func boolDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func validateTimeOfDay(field string, v *int) error {
	if v == nil {
		return nil
	}
	if *v < 0 || *v > TaskScheduleEndOfDay {
		return &ValidationError{Field: field, Message: field + " must be between 0 and 86400"}
	}
	return nil
}

func clampScheduleMax(n *int) int {
	if n == nil {
		return taskScheduleQueryDefaultMax
	}
	v := *n
	if v < taskScheduleQueryMinMax {
		return taskScheduleQueryMinMax
	}
	if v > taskScheduleQueryMaxMax {
		return taskScheduleQueryMaxMax
	}
	return v
}

// scheduleTimes reads the get_schedule action body.
//
// pyVergeOS prefers a "times" list, then a "schedule" list. An empty list is
// skipped, matching Python truthiness. A non-list value returns the whole
// object as the single entry. A top-level array is that list. An empty body
// is an empty list.
func scheduleTimes(raw json.RawMessage) ([]map[string]any, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return []map[string]any{}, nil
	}
	if raw[0] == '[' {
		var items []any
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, fmt.Errorf("vergeos: failed to decode schedule times: %w", err)
		}
		return objectList(items)
	}
	if raw[0] != '{' {
		return []map[string]any{}, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("vergeos: failed to decode schedule times: %w", err)
	}
	chosen, ok := firstTruthy(obj, "times", "schedule")
	if !ok {
		return []map[string]any{}, nil
	}
	items, ok := chosen.([]any)
	if !ok {
		return []map[string]any{obj}, nil
	}
	return objectList(items)
}

func firstTruthy(obj map[string]any, keys ...string) (any, bool) {
	for _, key := range keys {
		v, exists := obj[key]
		if exists && jsonTruthy(v) {
			return v, true
		}
	}
	return nil, false
}

func jsonTruthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0
	case []any:
		return len(x) != 0
	case map[string]any:
		return len(x) != 0
	default:
		return true
	}
}

func objectList(items []any) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("vergeos: schedule time entry is %T, want object", item)
		}
		out = append(out, obj)
	}
	return out, nil
}

// actionResult reads a row-action body. An empty body or a non-object is nil,
// which is what pyVergeOS returns when the action response is not a dict.
func actionResult(raw json.RawMessage) (map[string]any, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) || raw[0] != '{' {
		return nil, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("vergeos: failed to decode action response: %w", err)
	}
	return obj, nil
}

func putAction(ctx context.Context, client *Client, resource string, id int, endpoint string, body any) (map[string]any, error) {
	if body == nil {
		body = map[string]any{}
	}
	raw, err := client.putRaw(ctx, endpoint, body)
	if err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: resource, ID: id}
		}
		return nil, err
	}
	return actionResult(raw)
}

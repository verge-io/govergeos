package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// TaskScheduleTriggerService handles task_schedule_triggers.
type TaskScheduleTriggerService struct {
	client *Client
}

// List returns task schedule triggers.
func (s *TaskScheduleTriggerService) List(ctx context.Context, opts ...ListOption) ([]TaskScheduleTrigger, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = taskScheduleTriggerListFields
	}
	var rows []TaskScheduleTrigger
	if err := s.client.get(ctx, "/task_schedule_triggers", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByTask returns triggers for one task.
func (s *TaskScheduleTriggerService) ListByTask(ctx context.Context, taskID int, opts ...ListOption) ([]TaskScheduleTrigger, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("task eq %d", taskID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// ListBySchedule returns triggers for one schedule.
func (s *TaskScheduleTriggerService) ListBySchedule(ctx context.Context, scheduleID int, opts ...ListOption) ([]TaskScheduleTrigger, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("schedule eq %d", scheduleID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns one trigger by row key.
func (s *TaskScheduleTriggerService) Get(ctx context.Context, id int) (*TaskScheduleTrigger, error) {
	params := url.Values{}
	params.Set("fields", taskScheduleTriggerListFields)

	var row TaskScheduleTrigger
	endpoint := fmt.Sprintf("/task_schedule_triggers/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TaskScheduleTrigger", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

// Create links a task to a schedule and reads the trigger back.
func (s *TaskScheduleTriggerService) Create(ctx context.Context, req *TaskScheduleTriggerCreateRequest) (*TaskScheduleTrigger, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Task == 0 {
		return nil, &ValidationError{Field: "task", Message: "task is required"}
	}
	if req.Schedule == 0 {
		return nil, &ValidationError{Field: "schedule", Message: "schedule is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/task_schedule_triggers", req, &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete removes the link between a task and a schedule.
func (s *TaskScheduleTriggerService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/task_schedule_triggers/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "TaskScheduleTrigger", ID: id}
		}
		return err
	}
	return nil
}

// Trigger runs the linked task now.
//
// The call is PUT /task_schedule_triggers/{id}?action=trigger. The returned
// map is the action body. An empty body or a non-object body is nil.
func (s *TaskScheduleTriggerService) Trigger(ctx context.Context, id int) (map[string]any, error) {
	endpoint := fmt.Sprintf("/task_schedule_triggers/%d?action=trigger", id)
	return putAction(ctx, s.client, "TaskScheduleTrigger", id, endpoint, map[string]any{})
}

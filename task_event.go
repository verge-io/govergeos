package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// TaskEventService handles task_events.
type TaskEventService struct {
	client *Client
}

// List returns task events.
func (s *TaskEventService) List(ctx context.Context, opts ...ListOption) ([]TaskEvent, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = taskEventListFields
	}
	var rows []TaskEvent
	if err := s.client.get(ctx, "/task_events", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListByTask returns events linked to one task.
func (s *TaskEventService) ListByTask(ctx context.Context, taskID int, opts ...ListOption) ([]TaskEvent, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("task eq %d", taskID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// ListByOwner returns events for one owner key.
//
// The filter is "owner eq {id}", the same form pyVergeOS sends.
func (s *TaskEventService) ListByOwner(ctx context.Context, ownerID int, opts ...ListOption) ([]TaskEvent, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("owner eq %d", ownerID))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// ListByTable returns events whose source table matches.
func (s *TaskEventService) ListByTable(ctx context.Context, table string, opts ...ListOption) ([]TaskEvent, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("table eq '%s'", escapeFilterValue(table)))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// ListByEvent returns events with this event identifier.
func (s *TaskEventService) ListByEvent(ctx context.Context, event string, opts ...ListOption) ([]TaskEvent, error) {
	filterOpts := []ListOption{WithFilter(fmt.Sprintf("event eq '%s'", escapeFilterValue(event)))}
	filterOpts = append(filterOpts, opts...)
	return s.List(ctx, filterOpts...)
}

// Get returns one task event by row key.
func (s *TaskEventService) Get(ctx context.Context, id int) (*TaskEvent, error) {
	params := url.Values{}
	params.Set("fields", taskEventListFields)

	var row TaskEvent
	endpoint := fmt.Sprintf("/task_events/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "TaskEvent", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

// Create links a task to an event and reads the row back.
//
// The owner column is not sent. The platform fills it from the task.
func (s *TaskEventService) Create(ctx context.Context, req *TaskEventCreateRequest) (*TaskEvent, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Task == 0 {
		return nil, &ValidationError{Field: "task", Message: "task is required"}
	}
	if req.Table == "" {
		return nil, &ValidationError{Field: "table", Message: "table is required"}
	}
	if req.Event == "" {
		return nil, &ValidationError{Field: "event", Message: "event is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/task_events", req.wire(), &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update changes filters or context and reads the row back.
func (s *TaskEventService) Update(ctx context.Context, id int, req *TaskEventUpdateRequest) (*TaskEvent, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	endpoint := fmt.Sprintf("/task_events/%d", id)
	if err := s.client.put(ctx, endpoint, req.wire(), nil); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete removes the link between a task and an event.
func (s *TaskEventService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/task_events/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if statusNotFound(err) {
			return &NotFoundError{Resource: "TaskEvent", ID: id}
		}
		return err
	}
	return nil
}

// Trigger fires the event now.
//
// The call is PUT /task_events/{id}?action=trigger. eventContext nil sends
// an empty object. A non-nil map is sent as the context field. The returned
// map is the action body. An empty body or a non-object body is nil.
func (s *TaskEventService) Trigger(ctx context.Context, id int, eventContext map[string]any) (map[string]any, error) {
	body := map[string]any{}
	if eventContext != nil {
		body["context"] = eventContext
	}
	endpoint := fmt.Sprintf("/task_events/%d?action=trigger", id)
	return putAction(ctx, s.client, "TaskEvent", id, endpoint, body)
}

// taskEventWrite is the JSON for create and update. Map fields use any so an
// empty object is sent. omitempty on a map would drop that object.
type taskEventWrite struct {
	Task              int    `json:"task,omitempty"`
	Table             string `json:"table,omitempty"`
	Event             string `json:"event,omitempty"`
	EventName         string `json:"event_name,omitempty"`
	TableEventFilters any    `json:"table_event_filters,omitempty"`
	Context           any    `json:"context,omitempty"`
}

func (req *TaskEventCreateRequest) wire() taskEventWrite {
	body := taskEventWrite{
		Task:      req.Task,
		Table:     req.Table,
		Event:     req.Event,
		EventName: req.EventName,
	}
	if req.TableEventFilters != nil {
		body.TableEventFilters = req.TableEventFilters
	}
	if req.Context != nil {
		body.Context = req.Context
	}
	return body
}

func (req *TaskEventUpdateRequest) wire() taskEventWrite {
	var body taskEventWrite
	if req.TableEventFilters != nil {
		body.TableEventFilters = req.TableEventFilters
	}
	if req.Context != nil {
		body.Context = req.Context
	}
	return body
}

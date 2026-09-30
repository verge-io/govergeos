package vergeos

import (
	"context"
	"fmt"
	"net/url"
)

// TaskScriptService handles task_scripts.
type TaskScriptService struct {
	client *Client
}

// List returns task scripts.
func (s *TaskScriptService) List(ctx context.Context, opts ...ListOption) ([]TaskScript, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = taskScriptListFields
	}
	var rows []TaskScript
	if err := s.client.get(ctx, "/task_scripts", options.toQueryParams(), &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// Get returns one script by row key.
func (s *TaskScriptService) Get(ctx context.Context, id int) (*TaskScript, error) {
	params := url.Values{}
	params.Set("fields", taskScriptListFields)

	var row TaskScript
	endpoint := fmt.Sprintf("/task_scripts/%d", id)
	if err := s.client.get(ctx, endpoint, params, &row); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TaskScript", ID: id}
		}
		return nil, err
	}
	return &row, nil
}

// GetByName returns the script with this name.
//
// More than one row with the name is AmbiguousNameError.
func (s *TaskScriptService) GetByName(ctx context.Context, name string) (*TaskScript, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "TaskScript", ID: name}
	}
	if err := requireUniqueName("TaskScript", name, rows, func(row TaskScript) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("TaskScript", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(rows[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("TaskScript", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Create creates a script and reads it back.
//
// TaskSettings nil is sent as {"questions": []}.
func (s *TaskScriptService) Create(ctx context.Context, req *TaskScriptCreateRequest) (*TaskScript, error) {
	if req == nil {
		return nil, &ValidationError{Message: "create request is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}
	if req.Script == "" {
		return nil, &ValidationError{Field: "script", Message: "script is required"}
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/task_scripts", req.wire(), &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update updates a script and reads it back.
func (s *TaskScriptService) Update(ctx context.Context, id int, req *TaskScriptUpdateRequest) (*TaskScript, error) {
	if req == nil {
		return nil, &ValidationError{Message: "update request is required"}
	}
	endpoint := fmt.Sprintf("/task_scripts/%d", id)
	if err := s.client.put(ctx, endpoint, req.wire(), nil); err != nil {
		if IsNotFoundError(err) {
			return nil, &NotFoundError{Resource: "TaskScript", ID: id}
		}
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete deletes a script. The platform cascade-deletes tasks that use it.
func (s *TaskScriptService) Delete(ctx context.Context, id int) error {
	endpoint := fmt.Sprintf("/task_scripts/%d", id)
	if err := s.client.delete(ctx, endpoint); err != nil {
		if IsNotFoundError(err) {
			return &NotFoundError{Resource: "TaskScript", ID: id}
		}
		return err
	}
	return nil
}

// Run executes a script.
//
// The call is PUT /task_scripts/{id}?action=run. params nil sends an empty
// object. Otherwise the map is the action body, which is how script
// parameters are passed. The returned map is the action body. An empty body
// or a non-object body is nil.
func (s *TaskScriptService) Run(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	body := map[string]any{}
	if params != nil {
		body = params
	}
	endpoint := fmt.Sprintf("/task_scripts/%d?action=run", id)
	return putAction(ctx, s.client, "TaskScript", id, endpoint, body)
}

type taskScriptCreateBody struct {
	Name         string         `json:"name"`
	Script       string         `json:"script"`
	Description  string         `json:"description,omitempty"`
	TaskSettings map[string]any `json:"task_settings"`
}

type taskScriptUpdateBody struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Script       *string `json:"script,omitempty"`
	TaskSettings any     `json:"task_settings,omitempty"`
}

func (req *TaskScriptUpdateRequest) wire() taskScriptUpdateBody {
	body := taskScriptUpdateBody{
		Name:        req.Name,
		Description: req.Description,
		Script:      req.Script,
	}
	if req.TaskSettings != nil {
		body.TaskSettings = req.TaskSettings
	}
	return body
}

func (req *TaskScriptCreateRequest) wire() taskScriptCreateBody {
	settings := req.TaskSettings
	if settings == nil {
		settings = map[string]any{"questions": []any{}}
	}
	return taskScriptCreateBody{
		Name:         req.Name,
		Script:       req.Script,
		Description:  req.Description,
		TaskSettings: settings,
	}
}

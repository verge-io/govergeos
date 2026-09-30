package vergeos

import "encoding/json"

// JSONObject is a JSON object stored on a task-engine row.
//
// Null and a JSON value that is not an object decode as nil, so one odd
// cell does not reject the row. A string that itself contains an object
// is decoded. MarshalJSON sends the object.
type JSONObject map[string]any

// UnmarshalJSON implements json.Unmarshaler for JSONObject.
func (o *JSONObject) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*o = nil
		return nil
	}
	if data[0] == '{' {
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return err
		}
		*o = obj
		return nil
	}
	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil || encoded == "" || encoded[0] != '{' {
		*o = nil
		return nil
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(encoded), &obj); err != nil {
		*o = nil
		return nil
	}
	*o = obj
	return nil
}

// TaskEvent links a task to a system event on task_events.
//
// When the event occurs, the linked task runs. The owner column is filled
// by the platform from the parent task; create does not send it.
type TaskEvent struct {
	// Key is the event row key.
	Key FlexInt `json:"$key,omitempty"`
	// Owner is the owner reference, a "table/key" string or a bare key.
	Owner FlexString `json:"owner,omitempty"`
	// OwnerDisplay is the owner's display name.
	OwnerDisplay string `json:"owner_display,omitempty"`
	// Table is the event source table, such as alarms or vms.
	Table string `json:"table,omitempty"`
	// Event is the event identifier, such as login or lowered.
	Event string `json:"event,omitempty"`
	// EventName is the human-readable event name.
	EventName string `json:"event_name,omitempty"`
	// Task is the linked task.
	Task FlexFK `json:"task,omitempty"`
	// TaskName is the linked task's name.
	TaskName string `json:"task_name,omitempty"`
	// TaskDisplay is the linked task's display name.
	TaskDisplay string `json:"task_display,omitempty"`
	// TableEventFilters are the filter conditions stored on the row.
	TableEventFilters JSONObject `json:"table_event_filters,omitempty"`
	// Context is extra data passed to the task.
	Context JSONObject `json:"context,omitempty"`
	// Trigger is the timer value stored on the row.
	Trigger int `json:"trigger,omitempty"`
}

// TaskEventCreateRequest links a task to an event.
//
// Owner is not part of the request. The platform copies it from the task.
type TaskEventCreateRequest struct {
	// Task is the task row key (required).
	Task int `json:"task"`
	// Table is the event source table (required).
	Table string `json:"table"`
	// Event is the event identifier (required).
	Event string `json:"event"`
	// EventName is an optional human-readable name.
	EventName string `json:"event_name,omitempty"`
	// TableEventFilters are optional filter conditions.
	TableEventFilters map[string]any `json:"table_event_filters,omitempty"`
	// Context is optional data passed to the task.
	Context map[string]any `json:"context,omitempty"`
}

// TaskEventUpdateRequest changes filters or context on a task event.
//
// Nil fields are left unchanged. The task, table, and event are fixed after create.
type TaskEventUpdateRequest struct {
	TableEventFilters map[string]any `json:"table_event_filters,omitempty"`
	Context           map[string]any `json:"context,omitempty"`
}

// taskEventListFields is the projection pyVergeOS requests for task events.
const taskEventListFields = "$key,owner,owner#$display as owner_display,table,event,event_name,task,task#name as task_name,table_event_filters,trigger,context,task#$display as task_display"

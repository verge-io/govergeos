package vergeos

// TaskScript is GCS automation code on task_scripts.
type TaskScript struct {
	// Key is the script row key.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the script name.
	Name string `json:"name,omitempty"`
	// Description is an optional note.
	Description string `json:"description,omitempty"`
	// Script is the GCS source.
	Script string `json:"script,omitempty"`
	// TaskSettings holds the script's questions and other settings.
	TaskSettings JSONObject `json:"task_settings,omitempty"`
	// TaskCount is the number of tasks that use this script.
	TaskCount int `json:"task_count,omitempty"`
}

// TaskScriptCreateRequest is the body for creating a task script.
//
// TaskSettings nil sends {"questions": []}, which is the pyVergeOS default.
type TaskScriptCreateRequest struct {
	// Name is the script name (required).
	Name string `json:"name"`
	// Script is the GCS source (required).
	Script string `json:"script"`
	// Description is an optional note.
	Description string `json:"description,omitempty"`
	// TaskSettings are the script questions. Nil sends an empty question list.
	TaskSettings map[string]any `json:"task_settings,omitempty"`
}

// TaskScriptUpdateRequest is the body for updating a task script.
//
// Nil fields are left unchanged.
type TaskScriptUpdateRequest struct {
	Name         *string        `json:"name,omitempty"`
	Description  *string        `json:"description,omitempty"`
	Script       *string        `json:"script,omitempty"`
	TaskSettings map[string]any `json:"task_settings,omitempty"`
}

// taskScriptListFields is the projection pyVergeOS requests for scripts.
const taskScriptListFields = "$key,name,description,script,task_settings,count(tasks) as task_count"

package vergeos

// Repeat intervals for a task schedule. These are the values pyVergeOS sends
// as repeat_every.
const (
	TaskScheduleRepeatMinute = "minute"
	TaskScheduleRepeatHour   = "hour"
	TaskScheduleRepeatDay    = "day"
	TaskScheduleRepeatWeek   = "week"
	TaskScheduleRepeatMonth  = "month"
	TaskScheduleRepeatYear   = "year"
	TaskScheduleRepeatNever  = "never"
)

// Day-of-month settings for a monthly task schedule.
const (
	TaskScheduleDayOfMonthFirst     = "first"
	TaskScheduleDayOfMonthLast      = "last"
	TaskScheduleDayOfMonth15th      = "15th"
	TaskScheduleDayOfMonthStartDate = "start_date"
)

// Seconds in a day. The schedule end time defaults to this value, which is
// the last second boundary the task engine accepts for a time of day.
const TaskScheduleEndOfDay = 86400

// TaskSchedule is a reusable schedule on task_schedules.
//
// A schedule describes when work runs. TaskScheduleTriggers bind it to tasks.
type TaskSchedule struct {
	// Key is the schedule row key.
	Key FlexInt `json:"$key,omitempty"`
	// Name is the schedule name.
	Name string `json:"name,omitempty"`
	// Description is an optional note.
	Description string `json:"description,omitempty"`
	// Enabled reports whether the schedule is active.
	Enabled bool `json:"enabled,omitempty"`
	// Task is set when the schedule is bound to one task.
	Task FlexFK `json:"task,omitempty"`
	// TaskDisplay is the bound task's display name.
	TaskDisplay string `json:"task_display,omitempty"`
	// RepeatEvery is the interval: minute, hour, day, week, month, year, or never.
	RepeatEvery string `json:"repeat_every,omitempty"`
	// RepeatIteration is how many intervals pass between runs.
	RepeatIteration int `json:"repeat_iteration,omitempty"`
	// StartDate is the first day, YYYY-MM-DD.
	StartDate string `json:"start_date,omitempty"`
	// StartDateEpoch is the start date as a Unix timestamp.
	StartDateEpoch int64 `json:"start_date_epoch,omitempty"`
	// EndDate is the expiration, YYYY-MM-DD HH:MM:SS.
	EndDate string `json:"end_date,omitempty"`
	// StartTimeOfDay is seconds from midnight when the window opens.
	StartTimeOfDay int `json:"start_time_of_day,omitempty"`
	// EndTimeOfDay is seconds from midnight when the window closes.
	EndTimeOfDay int `json:"end_time_of_day,omitempty"`
	// AllDay reports a window that covers the whole day.
	AllDay bool `json:"all_day,omitempty"`
	// DayOfMonth is the monthly setting: first, last, 15th, or start_date.
	DayOfMonth string `json:"day_of_month,omitempty"`
	// Monday through Sunday select the days a weekly or daily schedule runs.
	Monday    bool `json:"monday,omitempty"`
	Tuesday   bool `json:"tuesday,omitempty"`
	Wednesday bool `json:"wednesday,omitempty"`
	Thursday  bool `json:"thursday,omitempty"`
	Friday    bool `json:"friday,omitempty"`
	Saturday  bool `json:"saturday,omitempty"`
	Sunday    bool `json:"sunday,omitempty"`
	// SystemCreated reports a schedule the platform created.
	SystemCreated bool `json:"system_created,omitempty"`
	// Creator is the user reference, a "table/key" string or a bare key.
	Creator FlexString `json:"creator,omitempty"`
	// CreatorDisplay is the creator's display name.
	CreatorDisplay string `json:"creator_display,omitempty"`
}

// TaskScheduleCreateRequest is the body for creating a task schedule.
//
// Nil pointers and empty strings take the same defaults pyVergeOS sends:
// enabled, every weekday, repeat_every hour, repeat_iteration 1,
// start_time_of_day 0, end_time_of_day 86400, and day_of_month start_date.
type TaskScheduleCreateRequest struct {
	// Name is the schedule name (required).
	Name string `json:"name"`
	// Description is an optional note.
	Description string `json:"description,omitempty"`
	// Enabled defaults to true when nil.
	Enabled *bool `json:"enabled,omitempty"`
	// RepeatEvery defaults to hour when empty.
	RepeatEvery string `json:"repeat_every,omitempty"`
	// RepeatIteration defaults to 1 when nil.
	RepeatIteration *int `json:"repeat_iteration,omitempty"`
	// StartDate is the first day, YYYY-MM-DD.
	StartDate string `json:"start_date,omitempty"`
	// EndDate is the expiration, YYYY-MM-DD HH:MM:SS.
	EndDate string `json:"end_date,omitempty"`
	// StartTimeOfDay is seconds from midnight. Nil sends 0.
	StartTimeOfDay *int `json:"start_time_of_day,omitempty"`
	// EndTimeOfDay is seconds from midnight. Nil sends 86400.
	EndTimeOfDay *int `json:"end_time_of_day,omitempty"`
	// DayOfMonth defaults to start_date when empty.
	DayOfMonth string `json:"day_of_month,omitempty"`
	// Monday through Sunday default to true when nil.
	Monday    *bool `json:"monday,omitempty"`
	Tuesday   *bool `json:"tuesday,omitempty"`
	Wednesday *bool `json:"wednesday,omitempty"`
	Thursday  *bool `json:"thursday,omitempty"`
	Friday    *bool `json:"friday,omitempty"`
	Saturday  *bool `json:"saturday,omitempty"`
	Sunday    *bool `json:"sunday,omitempty"`
	// Task binds the schedule to one task.
	Task *int `json:"task,omitempty"`
}

// TaskScheduleUpdateRequest is the body for updating a task schedule.
//
// Nil fields are left unchanged. The task engine does not accept task on update.
type TaskScheduleUpdateRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
	RepeatEvery     *string `json:"repeat_every,omitempty"`
	RepeatIteration *int    `json:"repeat_iteration,omitempty"`
	StartDate       *string `json:"start_date,omitempty"`
	EndDate         *string `json:"end_date,omitempty"`
	StartTimeOfDay  *int    `json:"start_time_of_day,omitempty"`
	EndTimeOfDay    *int    `json:"end_time_of_day,omitempty"`
	DayOfMonth      *string `json:"day_of_month,omitempty"`
	Monday          *bool   `json:"monday,omitempty"`
	Tuesday         *bool   `json:"tuesday,omitempty"`
	Wednesday       *bool   `json:"wednesday,omitempty"`
	Thursday        *bool   `json:"thursday,omitempty"`
	Friday          *bool   `json:"friday,omitempty"`
	Saturday        *bool   `json:"saturday,omitempty"`
	Sunday          *bool   `json:"sunday,omitempty"`
}

// TaskScheduleQuery selects upcoming runs from the get_schedule action.
//
// MaxResults nil uses 100. A set value is clamped to 1 through 1440, which
// is the range the task engine accepts.
type TaskScheduleQuery struct {
	MaxResults *int `json:"-"`
	// StartTime and EndTime are Unix timestamps that bound the window.
	StartTime *int64 `json:"start_time,omitempty"`
	EndTime   *int64 `json:"end_time,omitempty"`
}

// taskScheduleListFields is the projection pyVergeOS requests for schedules.
const taskScheduleListFields = "$key,name,description,enabled,task,task#$display as task_display,repeat_every,repeat_iteration,start_date,start_date_epoch,end_date,start_time_of_day,end_time_of_day,all_day,day_of_month,monday,tuesday,wednesday,thursday,friday,saturday,sunday,system_created,creator,creator#$display as creator_display"

const (
	taskScheduleQueryDefaultMax = 100
	taskScheduleQueryMinMax     = 1
	taskScheduleQueryMaxMax     = 1440
)

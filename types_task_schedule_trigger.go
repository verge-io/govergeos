package vergeos

// TaskScheduleTrigger links a task to a schedule on task_schedule_triggers.
//
// When the schedule fires, every task linked through a trigger runs.
type TaskScheduleTrigger struct {
	// Key is the trigger row key.
	Key FlexInt `json:"$key,omitempty"`
	// Task is the linked task.
	Task FlexFK `json:"task,omitempty"`
	// TaskDisplay is the linked task's display name.
	TaskDisplay string `json:"task_display,omitempty"`
	// Schedule is the linked schedule.
	Schedule FlexFK `json:"schedule,omitempty"`
	// ScheduleDisplay is the linked schedule's display name.
	ScheduleDisplay string `json:"schedule_display,omitempty"`
	// Trigger is the timer value stored on the row.
	Trigger int `json:"trigger,omitempty"`
	// ScheduleKey is the flattened schedule row key (sch_$key).
	ScheduleKey FlexFK `json:"sch_$key,omitempty"`
	// ScheduleEnabled is the flattened schedule enabled flag.
	ScheduleEnabled bool `json:"sch_enabled,omitempty"`
	// ScheduleStartTimeOfDay is the flattened start time, in seconds from midnight.
	ScheduleStartTimeOfDay int `json:"sch_start_time_of_day,omitempty"`
	// ScheduleRepeatIteration is the flattened repeat count.
	ScheduleRepeatIteration int `json:"sch_repeat_iteration,omitempty"`
	// ScheduleEndDate is the flattened expiration.
	ScheduleEndDate string `json:"sch_end_date,omitempty"`
	// ScheduleDayOfMonth is the flattened day-of-month display.
	ScheduleDayOfMonth string `json:"sch_day_of_month,omitempty"`
	// ScheduleRepeatEvery is the flattened repeat interval display.
	ScheduleRepeatEvery string `json:"sch_repeat_every,omitempty"`
}

// TaskScheduleTriggerCreateRequest links a task to a schedule.
type TaskScheduleTriggerCreateRequest struct {
	// Task is the task row key (required).
	Task int `json:"task"`
	// Schedule is the schedule row key (required).
	Schedule int `json:"schedule"`
}

// taskScheduleTriggerListFields is the projection pyVergeOS requests for triggers.
const taskScheduleTriggerListFields = "$key,task,schedule,trigger," +
	"flatten(schedule[$key as sch_$key,enabled as sch_enabled,start_time_of_day as sch_start_time_of_day,repeat_iteration as sch_repeat_iteration,end_date as sch_end_date,display(day_of_month) as sch_day_of_month,display(repeat_every) as sch_repeat_every])," +
	"display(task) as task_display,display(schedule) as schedule_display"

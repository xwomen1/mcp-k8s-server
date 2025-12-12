// internal/domain/job.go
package domain

import "time"

type JobName string

type JobStatus string

const (
	JobStatusActive    JobStatus = "Active"
	JobStatusSucceeded JobStatus = "Succeeded"
	JobStatusFailed    JobStatus = "Failed"
	JobStatusComplete  JobStatus = "Complete"
)

type JobConditionType string

const (
	JobComplete JobConditionType = "Complete"
	JobFailed   JobConditionType = "Failed"
)

type Job struct {
	Name           JobName           `json:"name"`
	Namespace      Namespace         `json:"namespace"`
	Status         JobStatus         `json:"status"`
	Active         int32             `json:"active"`
	Succeeded      int32             `json:"succeeded"`
	Failed         int32             `json:"failed"`
	Completions    *int32            `json:"completions,omitempty"`
	Parallelism    *int32            `json:"parallelism,omitempty"`
	BackoffLimit   *int32            `json:"backoff_limit,omitempty"`
	StartTime      *time.Time        `json:"start_time,omitempty"`
	CompletionTime *time.Time        `json:"completion_time,omitempty"`
	Duration       string            `json:"duration,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

type JobCreateOptions struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Image         string            `json:"image"`
	Command       []string          `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	Completions   *int32            `json:"completions,omitempty"`
	Parallelism   *int32            `json:"parallelism,omitempty"`
	BackoffLimit  *int32            `json:"backoff_limit,omitempty"`
	RestartPolicy string            `json:"restart_policy,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

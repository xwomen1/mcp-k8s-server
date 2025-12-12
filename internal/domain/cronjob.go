// internal/domain/cronjob.go
package domain

import "time"

type CronJobName string

type CronJobStatus string

const (
	CronJobStatusActive    CronJobStatus = "Active"
	CronJobStatusSuspended CronJobStatus = "Suspended"
)

type CronJob struct {
	Name                       CronJobName       `json:"name"`
	Namespace                  Namespace         `json:"namespace"`
	Schedule                   string            `json:"schedule"`
	Suspend                    bool              `json:"suspend"`
	Active                     int               `json:"active"`
	LastScheduleTime           *time.Time        `json:"last_schedule_time,omitempty"`
	SuccessfulJobsHistoryLimit *int32            `json:"successful_jobs_history_limit,omitempty"`
	FailedJobsHistoryLimit     *int32            `json:"failed_jobs_history_limit,omitempty"`
	ConcurrencyPolicy          string            `json:"concurrency_policy"`
	Status                     CronJobStatus     `json:"status"`
	Labels                     map[string]string `json:"labels,omitempty"`
	CreatedAt                  time.Time         `json:"created_at"`
}

type CronJobCreateOptions struct {
	Name                       string            `json:"name"`
	Namespace                  string            `json:"namespace"`
	Schedule                   string            `json:"schedule"`
	Image                      string            `json:"image"`
	Command                    []string          `json:"command,omitempty"`
	Args                       []string          `json:"args,omitempty"`
	Suspend                    bool              `json:"suspend,omitempty"`
	ConcurrencyPolicy          string            `json:"concurrency_policy,omitempty"`
	SuccessfulJobsHistoryLimit *int32            `json:"successful_jobs_history_limit,omitempty"`
	FailedJobsHistoryLimit     *int32            `json:"failed_jobs_history_limit,omitempty"`
	RestartPolicy              string            `json:"restart_policy,omitempty"`
	Labels                     map[string]string `json:"labels,omitempty"`
	Env                        map[string]string `json:"env,omitempty"`
}

package models

import (
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID               string
	TraceID          string
	OriginalFilename string
	FilePath         string
	Status           TaskStatus
	ErrorMessage     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      *time.Time
}

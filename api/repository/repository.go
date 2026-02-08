package repository

import (
	"context"
	"errors"

	"mediaConverter/api/models"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
)

type Repository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	GetTaskByTraceID(ctx context.Context, traceID string) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, id string, status models.TaskStatus, errorMessage string) error
}

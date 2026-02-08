package service

import (
	"context"

	"mediaConverter/api/cache"
	"mediaConverter/api/dto"
	"mediaConverter/api/kafka"
	"mediaConverter/api/models"
	"mediaConverter/api/repository"
)

type TaskService struct {
	repo     repository.Repository
	cache    *cache.StatusCache
	producer kafka.Producer
	topic    string
}

func NewTaskService(repo repository.Repository, cache *cache.StatusCache, producer kafka.Producer) *TaskService {
	return &TaskService{
		repo:     repo,
		cache:    cache,
		producer: producer,
		topic:    "media_tasks",
	}
}

func (s *TaskService) CreateTask(ctx context.Context, traceID string, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	task := &models.Task{
		TraceID:          traceID,
		OriginalFilename: req.OriginalFilename,
		FilePath:         req.FilePath,
		OutputFormat:     req.OutputFormat,
		TargetWidth:      req.TargetWidth,
		TargetHeight:     req.TargetHeight,
		Crop:             req.Crop,
		Status:           models.StatusPending,
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	s.cache.Set(ctx, task.ID, models.StatusPending)

	msg := &kafka.TaskMessage{
		TaskID:       task.ID,
		TraceID:      traceID,
		FilePath:     req.FilePath,
		OutputFormat: req.OutputFormat,
		TargetWidth:  req.TargetWidth,
		TargetHeight: req.TargetHeight,
		Crop:         req.Crop,
	}
	if err := s.producer.SendTaskMessage(ctx, s.topic, msg); err != nil {
		return nil, err
	}

	return s.toResponse(task), nil
}

func (s *TaskService) GetTaskStatus(ctx context.Context, taskID string) (*dto.TaskResponse, error) {
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return s.toResponse(task), nil
}

func (s *TaskService) toResponse(task *models.Task) *dto.TaskResponse {
	var completedAt *string
	if task.CompletedAt != nil {
		formatted := task.CompletedAt.Format("2006-01-02T15:04:05Z")
		completedAt = &formatted
	}

	var outputFilename string
	if task.Status == models.StatusCompleted {
		ext := task.OutputFormat
		if ext == "" {
			ext = "jpg"
		}
		outputFilename = task.ID + "." + ext
	}

	return &dto.TaskResponse{
		ID:               task.ID,
		TraceID:          task.TraceID,
		OriginalFilename: task.OriginalFilename,
		OutputFilename:   outputFilename,
		OutputFormat:     task.OutputFormat,
		TargetWidth:      task.TargetWidth,
		TargetHeight:     task.TargetHeight,
		Crop:             task.Crop,
		Status:           string(task.Status),
		ErrorMessage:     task.ErrorMessage,
		CreatedAt:        task.CreatedAt.Format("2006-01-02T15:04:05Z"),
		CompletedAt:      completedAt,
	}
}

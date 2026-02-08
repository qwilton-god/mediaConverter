package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"mediaConverter/api/database"
	"mediaConverter/api/models"
)

type PostgresRepo struct {
	db *database.DB
}

func NewPostgresRepo(db *database.DB) Repository {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) CreateTask(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (trace_id, original_filename, file_path, status, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	var createdTask models.Task
	err := r.db.Pool.QueryRow(ctx, query,
		task.TraceID,
		task.OriginalFilename,
		task.FilePath,
		task.Status,
		task.ErrorMessage,
	).Scan(&createdTask.ID, &createdTask.CreatedAt, &createdTask.UpdatedAt)

	if err != nil {
		return err
	}

	task.ID = createdTask.ID
	task.CreatedAt = createdTask.CreatedAt
	task.UpdatedAt = createdTask.UpdatedAt

	return nil
}

func (r *PostgresRepo) GetTask(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, trace_id, original_filename, file_path, status, error_message, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, id)

	var task models.Task
	err := row.Scan(
		&task.ID,
		&task.TraceID,
		&task.OriginalFilename,
		&task.FilePath,
		&task.Status,
		&task.ErrorMessage,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (r *PostgresRepo) GetTaskByTraceID(ctx context.Context, traceID string) (*models.Task, error) {
	query := `
		SELECT id, trace_id, original_filename, file_path, status, error_message, created_at, updated_at, completed_at
		FROM tasks
		WHERE trace_id = $1
	`

	row := r.db.Pool.QueryRow(ctx, query, traceID)

	var task models.Task
	err := row.Scan(
		&task.ID,
		&task.TraceID,
		&task.OriginalFilename,
		&task.FilePath,
		&task.Status,
		&task.ErrorMessage,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (r *PostgresRepo) UpdateTaskStatus(ctx context.Context, id string, status models.TaskStatus, errorMessage string) error {
	query := `
		UPDATE tasks
		SET status = $1, error_message = $2, updated_at = NOW()
	`

	if status == models.StatusCompleted || status == models.StatusFailed {
		query += `, completed_at = NOW()`
	}

	query += ` WHERE id = $3`

	result, err := r.db.Pool.Exec(ctx, query, status, errorMessage, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}

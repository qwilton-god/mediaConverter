package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	UpdateTaskStatus(ctx context.Context, taskID string, status string, errMsg string) error
}

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) UpdateTaskStatus(ctx context.Context, taskID string, status string, errMsg string) error {
	query := `UPDATE tasks SET status = $1, error_message = $2, updated_at = NOW()`
	if status == "completed" || status == "failed" {
		query += `, completed_at = NOW()`
	}
	query += ` WHERE id = $3`

	_, err := r.db.Exec(ctx, query, status, errMsg, taskID)
	return err
}

var ErrTaskNotFound = errors.New("task not found")

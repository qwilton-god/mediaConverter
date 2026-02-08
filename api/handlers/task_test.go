package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"

	"mediaConverter/api/dto"
	"mediaConverter/api/middleware"
	"mediaConverter/api/models"
)

type mockTaskService struct {
	createTaskFunc func(ctx context.Context, traceID string, req *dto.CreateTaskRequest) (*dto.TaskResponse, error)
	getTaskFunc    func(ctx context.Context, taskID string) (*dto.TaskResponse, error)
}

func (m *mockTaskService) CreateTask(ctx context.Context, traceID string, req *dto.CreateTaskRequest) (*dto.TaskResponse, error) {
	if m.createTaskFunc != nil {
		return m.createTaskFunc(ctx, traceID, req)
	}
	taskID := uuid.New().String()
	return &dto.TaskResponse{
		ID:               taskID,
		TraceID:          traceID,
		OriginalFilename: req.OriginalFilename,
		Status:           string(models.StatusPending),
		CreatedAt:        time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (m *mockTaskService) GetTaskStatus(ctx context.Context, taskID string) (*dto.TaskResponse, error) {
	if m.getTaskFunc != nil {
		return m.getTaskFunc(ctx, taskID)
	}
	return &dto.TaskResponse{
		ID:        taskID,
		TraceID:   uuid.New().String(),
		Status:    string(models.StatusCompleted),
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}, nil
}

func createTestImageFile(t *testing.T) (*os.File, *multipart.FileHeader) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.jpg")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	content := make([]byte, 1024)
	copy(content, []byte{0xFF, 0xD8, 0xFF, 0xE0})
	if _, err := file.Write(content); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close test file: %v", err)
	}

	file, err = os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	header := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     1024,
	}

	return file, header
}

func TestTaskHandler_Upload_Success(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping test: requires root to create /uploads directory")
	}

	uploadsDir := "/uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatalf("Failed to create uploads dir: %v", err)
	}
	defer os.RemoveAll(uploadsDir)

	logger := zaptest.NewLogger(t)
	mockService := &mockTaskService{}
	handler := NewTaskHandler(mockService, logger)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(testFilePath, []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	fileContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		t.Fatalf("Failed to write form file: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	traceID := uuid.New().String()
	req.Header.Set("X-Trace-ID", traceID)

	ctx := context.WithValue(req.Context(), middleware.TraceIDKey, traceID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestTaskHandler_Upload_NoFile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := &mockTaskService{}
	handler := NewTaskHandler(mockService, logger)

	req := httptest.NewRequest("POST", "/upload", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data")
	traceID := uuid.New().String()
	req.Header.Set("X-Trace-ID", traceID)

	ctx := context.WithValue(req.Context(), middleware.TraceIDKey, traceID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Upload(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestTaskHandler_Status_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	taskID := uuid.New().String()
	traceID := uuid.New().String()

	mockService := &mockTaskService{
		getTaskFunc: func(ctx context.Context, id string) (*dto.TaskResponse, error) {
			return &dto.TaskResponse{
				ID:        taskID,
				TraceID:   traceID,
				Status:    string(models.StatusCompleted),
				CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
			}, nil
		},
	}

	handler := NewTaskHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/status/"+taskID, nil)
	req.Header.Set("X-Trace-ID", traceID)

	ctx := context.WithValue(req.Context(), middleware.TraceIDKey, traceID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Status(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestTaskHandler_Status_NotFound(t *testing.T) {
	logger := zaptest.NewLogger(t)
	taskID := uuid.New().String()
	traceID := uuid.New().String()

	mockService := &mockTaskService{
		getTaskFunc: func(ctx context.Context, id string) (*dto.TaskResponse, error) {
			return nil, dto.ErrTaskNotFound
		},
	}

	handler := NewTaskHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/status/"+taskID, nil)
	req.Header.Set("X-Trace-ID", traceID)

	ctx := context.WithValue(req.Context(), middleware.TraceIDKey, traceID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Status(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}
}

func TestTaskHandler_Status_EmptyTaskID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockService := &mockTaskService{}
	handler := NewTaskHandler(mockService, logger)

	req := httptest.NewRequest("GET", "/status/", nil)
	traceID := uuid.New().String()
	req.Header.Set("X-Trace-ID", traceID)

	ctx := context.WithValue(req.Context(), middleware.TraceIDKey, traceID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.Status(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

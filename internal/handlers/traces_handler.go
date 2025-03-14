package handlers

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type TraceHandler struct {
	Store      *store.Storage
	BucketName string
}

func NewTraceHandler(store *store.Storage, bucketName string) *TraceHandler {
	return &TraceHandler{
		Store:      store,
		BucketName: bucketName,
	}
}

func (h *TraceHandler) GetTraceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courseID := chi.URLParam(r, "course_id")
	traceID := chi.URLParam(r, "trace_id")

	trace, err := h.Store.Traces.GetTraceByID(r.Context(), courseID, traceID)
	if err != nil {
		http.Error(w, `{"error": "Trace not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(trace)
}

func (h *TraceHandler) GetAllTracesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courseID := chi.URLParam(r, "course_id")

	traces, err := h.Store.Traces.GetTracesByCourseID(r.Context(), courseID)
	if err != nil {
		http.Error(w, `{"error": "Could not fetch traces"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(traces)
}

func (h *TraceHandler) DeleteTraceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courseID := chi.URLParam(r, "course_id")
	traceID := chi.URLParam(r, "trace_id")

	// ✅ Step 1: Check if the trace exists
	trace, err := h.Store.Traces.GetTraceByID(r.Context(), courseID, traceID)
	if err != nil {
		http.Error(w, `{"error": "Trace not found"}`, http.StatusNotFound)
		return
	}

	// ✅ Step 2: Extract file path from GCS URL
	fileName := extractFileNameFromURL(trace.BucketPath)

	// ✅ Step 3: Delete file from GCS
	err = h.deleteFileFromGCS(r.Context(), fileName)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete file from GCS"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Step 4: Delete trace from database
	if err := h.Store.Traces.DeleteTrace(r.Context(), courseID, traceID); err != nil {
		http.Error(w, `{"error": "Could not delete trace from database"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Step 5: Respond with success
	w.WriteHeader(http.StatusNoContent)
}

func (h *TraceHandler) UploadTraceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse course ID from URL
	courseIDStr := chi.URLParam(r, "course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error": "Invalid course ID"}`, http.StatusBadRequest)
		return
	}

	// ✅ Extract Basic Auth for `user_id`
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Unauthorized: Missing Basic Auth"}`, http.StatusUnauthorized)
		return
	}

	// ✅ Get user by credentials
	user, err := h.Store.Users.GetUserByCredentials(r.Context(), username, password)
	if err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Parse file from form-data
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, `{"error": "Failed to parse multipart form"}`, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error": "Missing file in request"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique filename using timestamp
	uniqueFilename := fmt.Sprintf("%d-%s", time.Now().Unix(), handler.Filename)

	// Upload file to GCS
	gcsURL, err := h.uploadFileToGCS(r.Context(), file, uniqueFilename)
	if err != nil {
		http.Error(w, `{"error": "Failed to upload file to GCS"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Store metadata in database with `user_id`
	trace := &store.Trace{
		CourseID:    uint(courseID),
		UserID:      user.ID, // ✅ Add user_id from authentication
		FileName:    handler.Filename,
		BucketPath:  gcsURL,
		DateCreated: time.Now(),
	}

	if err := h.Store.Traces.CreateTrace(r.Context(), trace); err != nil {
		http.Error(w, `{"error": "Could not save trace metadata"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Respond with uploaded file metadata
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trace)
}

func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1] // Extract last part of URL as filename
}

func (h *TraceHandler) uploadFileToGCS(ctx context.Context, file io.Reader, fileName string) (string, error) {
	// Log the environment variable path
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	fmt.Println("Using credentials file:", credFile)

	// Ensure the credentials file exists
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		fmt.Println("ERROR: Google Cloud credentials file does not exist.")
		return "", fmt.Errorf("google cloud credentials file not found")
	}

	// Create GCS client
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credFile))
	if err != nil {
		fmt.Println("ERROR: Failed to create GCS client:", err)
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	// Log bucket and file name
	fmt.Printf("Uploading file '%s' to bucket '%s'\n", fileName, h.BucketName)

	bucket := client.Bucket(h.BucketName)
	object := bucket.Object(fileName)
	writer := object.NewWriter(ctx)
	writer.ContentType = "application/octet-stream"

	// Copy file to GCS
	if _, err := io.Copy(writer, file); err != nil {
		fmt.Println("ERROR: Failed to upload file to GCS:", err)
		return "", fmt.Errorf("failed to upload file to GCS: %w", err)
	}

	// Finalize the upload
	if err := writer.Close(); err != nil {
		fmt.Println("ERROR: Failed to finalize GCS upload:", err)
		return "", fmt.Errorf("failed to finalize GCS upload: %w", err)
	}

	// Generate GCS URL
	fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.BucketName, fileName)
	fmt.Println("File successfully uploaded:", fileURL)

	return fileURL, nil
}

func (h *TraceHandler) deleteFileFromGCS(ctx context.Context, fileName string) error {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(h.BucketName)
	object := bucket.Object(fileName)

	// Delete the file from GCS
	err = object.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete file from GCS: %w", err)
	}

	fmt.Println("File deleted successfully from GCS:", fileName)
	return nil
}

package handlers

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
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

func extractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1] // Extract last part of URL as filename
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

	courseIDStr := chi.URLParam(r, "course_id")
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		http.Error(w, `{"error": "Invalid course ID"}`, http.StatusBadRequest)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Unauthorized: Missing Basic Auth"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.Store.Users.GetUserByCredentials(r.Context(), username, password)
	if err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	err = r.ParseMultipartForm(50 << 20) // support larger payload
	if err != nil {
		http.Error(w, `{"error": "Failed to parse multipart form"}`, http.StatusBadRequest)
		return
	}

	formFiles := r.MultipartForm.File["files"] // same key as Postman

	var uploadedTraces []*store.Trace

	for _, fileHeader := range formFiles {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, `{"error": "Failed to open file"}`, http.StatusInternalServerError)
			return
		}
		defer file.Close()

		uniqueFilename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), fileHeader.Filename)

		gcsURL, err := h.uploadFileToGCS(r.Context(), file, uniqueFilename)
		if err != nil {
			http.Error(w, `{"error": "Failed to upload file to GCS"}`, http.StatusInternalServerError)
			return
		}

		trace := &store.Trace{
			CourseID:    uint(courseID),
			UserID:      user.ID,
			FileName:    fileHeader.Filename,
			BucketPath:  gcsURL,
			DateCreated: time.Now(),
		}

		if err := h.Store.Traces.CreateTrace(r.Context(), trace); err != nil {
			http.Error(w, `{"error": "Could not save trace metadata"}`, http.StatusInternalServerError)
			return
		}

		uploadedTraces = append(uploadedTraces, trace)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(uploadedTraces)
}

func (h *TraceHandler) uploadFileToGCS(ctx context.Context, file io.Reader, fileName string) (string, error) {
	// Log the environment variable path

	fmt.Println("Using service account")

	// Create GCS client
	client, err := storage.NewClient(ctx)
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

	fmt.Println("Using service account")
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println("Error", err)
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

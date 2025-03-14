package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type InstructorHandler struct {
	Store *store.Storage
}

func NewInstructorHandler(store *store.Storage) *InstructorHandler {
	return &InstructorHandler{Store: store}
}

func (h *InstructorHandler) CreateInstructorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 🔹 Extract Basic Auth Credentials
	username, _, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Missing basic auth credentials"}`, http.StatusUnauthorized)
		return
	}

	// 🔹 Parse JSON Request Body
	var data struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// 🔹 Create Instructor Struct
	instructor := store.Instructor{
		Name: data.Name,
	}

	// 🔹 Call `CreateInstructor` with `username`
	if err := h.Store.Instructors.CreateInstructor(r.Context(), username, &instructor); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 🔹 Return Created Instructor
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(instructor)
}

func (h *InstructorHandler) GetInstructorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	instructorID := chi.URLParam(r, "instructorId")

	instructor, err := h.Store.Instructors.GetInstructorByID(r.Context(), instructorID)
	if err != nil {
		http.Error(w, `{"error": "Instructor not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(instructor)
}

func (h *InstructorHandler) UpdateInstructorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	instructorID := chi.URLParam(r, "instructorId")

	var updateData store.Instructor
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.Store.Instructors.UpdateInstructor(r.Context(), instructorID, &updateData); err != nil {
		http.Error(w, "Could not update instructor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateData)
}

func (h *InstructorHandler) DeleteInstructorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	instructorID := chi.URLParam(r, "instructorId")

	// ✅ Check if the instructor exists before deleting
	instructor, err := h.Store.Instructors.GetInstructorByID(r.Context(), instructorID)
	if err != nil {
		http.Error(w, `{"error": "Instructor not found"}`, http.StatusNotFound)
		return
	}

	// ✅ Proceed with deletion if the instructor exists
	if err := h.Store.Instructors.DeleteInstructor(r.Context(), instructorID); err != nil {
		http.Error(w, `{"error": "Could not delete instructor"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Return success response with deleted instructor details
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Instructor deleted successfully",
		"instructorID": instructorID,
		"name":         instructor.Name,
	})
}

func (h *InstructorHandler) PatchInstructorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	instructorID := chi.URLParam(r, "instructorId")

	// ✅ Check if the instructor exists
	existingInstructor, err := h.Store.Instructors.GetInstructorByID(r.Context(), instructorID)
	if err != nil {
		http.Error(w, `{"error": "Instructor not found"}`, http.StatusNotFound)
		return
	}

	// ✅ Parse JSON request body
	var updateData store.Instructor
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// ✅ Perform partial update only for provided fields
	if updateData.Name != "" {
		existingInstructor.Name = updateData.Name
	}

	// ✅ Call the update function
	if err := h.Store.Instructors.UpdateInstructor(r.Context(), instructorID, existingInstructor); err != nil {
		http.Error(w, `{"error": "Could not update instructor"}`, http.StatusInternalServerError)
		return
	}

	// ✅ Return updated instructor
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingInstructor)
}

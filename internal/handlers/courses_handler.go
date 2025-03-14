package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type CourseHandler struct {
	Store *store.Storage
}

func NewCourseHandler(store *store.Storage) *CourseHandler {
	return &CourseHandler{Store: store}
}

func (h *CourseHandler) CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 🔹 Step 1: Extract Basic Auth
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, `{"error": "Unauthorized. Missing Basic Auth"}`, http.StatusUnauthorized)
		return
	}

	// 🔹 Step 2: Get user ID from username & password
	user, err := h.Store.Users.GetUserByCredentials(r.Context(), username, password)
	if err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// 🔹 Step 3: Parse JSON request body
	var course store.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// 🔹 Step 4: Assign authenticated user as `owner_user_id`
	course.OwnerUserID = user.ID

	// 🔹 Debugging Logs
	fmt.Printf("Received instructor_id: %d\n", course.InstructorID)
	fmt.Printf("Authenticated owner_user_id: %d\n", course.OwnerUserID)

	// 🔹 Step 5: Ensure instructor exists
	if err := h.Store.Instructors.CheckInstructorExists(r.Context(), course.InstructorID); err != nil {
		http.Error(w, `{"error": "Instructor does not exist"}`, http.StatusBadRequest)
		return
	}

	// 🔹 Step 6: Create Course
	if err := h.Store.Courses.CreateCourse(r.Context(), &course); err != nil {
		http.Error(w, `{"error": "Could not create course"}`, http.StatusInternalServerError)
		return
	}

	// 🔹 Step 7: Respond with Created Course
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (h *CourseHandler) GetCourseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courseID, err := strconv.Atoi(chi.URLParam(r, "courseId"))
	if err != nil {
		http.Error(w, `{"error": "Invalid course ID"}`, http.StatusBadRequest)
		return
	}

	course, err := h.Store.Courses.GetCourseByID(r.Context(), uint(courseID))
	if err != nil {
		http.Error(w, `{"error": "Course not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(course)
}

func (h *CourseHandler) UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courseID, err := strconv.Atoi(chi.URLParam(r, "courseId"))
	if err != nil {
		http.Error(w, `{"error": "Invalid course ID"}`, http.StatusBadRequest)
		return
	}

	var updateData store.Course
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.Store.Courses.UpdateCourse(r.Context(), uint(courseID), &updateData); err != nil {
		http.Error(w, `{"error": "Could not update course"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateData)
}

func (h *CourseHandler) PatchCourseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 🔹 Extract course ID from URL
	courseID, err := strconv.Atoi(chi.URLParam(r, "courseId"))
	if err != nil {
		http.Error(w, `{"error": "Invalid course ID"}`, http.StatusBadRequest)
		return
	}

	// 🔹 Check if the course exists
	_, err = h.Store.Courses.GetCourseByID(r.Context(), uint(courseID))
	if err != nil {
		http.Error(w, `{"error": "Course not found"}`, http.StatusNotFound)
		return
	}

	// 🔹 Decode JSON request payload
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// 🔹 Perform the update
	if err := h.Store.Courses.PatchCourse(r.Context(), uint(courseID), updateData); err != nil {
		http.Error(w, `{"error": "Could not patch course"}`, http.StatusInternalServerError)
		return
	}

	// 🔹 Respond with updated data
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateData)
}

func (h *CourseHandler) DeleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	courseID, err := strconv.Atoi(chi.URLParam(r, "courseId"))
	if err != nil {
		http.Error(w, `{"error": "Invalid course ID"}`, http.StatusBadRequest)
		return
	}

	// 🔹 Step 1: Check if course exists
	existingCourse, err := h.Store.Courses.GetCourseByID(r.Context(), uint(courseID))
	if err != nil {
		http.Error(w, `{"error": "Course not found"}`, http.StatusNotFound)
		return
	}

	// 🔹 Step 2: Delete the course
	if err := h.Store.Courses.DeleteCourse(r.Context(), existingCourse.ID); err != nil {
		http.Error(w, `{"error": "Could not delete course"}`, http.StatusInternalServerError)
		return
	}

	// 🔹 Step 3: Respond with success
	w.WriteHeader(http.StatusNoContent)
}

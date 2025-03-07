package main

import (
	"encoding/json"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"strconv"
	"time"
)

func NewApplication(storage *store.Storage) *application {
	return &application{
		config: config{addr: ":8080"},
		store:  storage,
	}
}

type application struct {
	config config
	store  *store.Storage
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func (a *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", a.healthCheckHandler)
		r.Get("/users/{id}", a.getUserHandler)
		r.Post("/users", a.createUserHandler)
		r.Put("/users/{id}", a.updateUserHandler)
		r.Delete("/users/{id}", a.deleteUserHandler)

		r.Get("/courses", a.getCoursesHandler)
		r.Post("/courses", a.createCourseHandler)

		r.Get("/instructors", a.getInstructorsHandler)
		r.Post("/instructors", a.createInstructorHandler)

		r.Get("/trace/{id}", a.getTraceHandler)
		r.Post("/trace/upload", a.uploadTraceHandler)
	})
	return r
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user store.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Store user in DB
	if err := app.store.Users.CreateUser(r.Context(), &user); err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := app.store.Users.GetUserByID(r.Context(), uint(id))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (a *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         a.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}
	log.Println("Starting server on", a.config.addr)
	return srv.ListenAndServe()
}

func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updateData store.User
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update the user in DB
	if err := app.store.Users.UpdateUser(r.Context(), uint(id), &updateData); err != nil {
		http.Error(w, "Could not update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateData)
}

func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Delete user from DB
	if err := app.store.Users.DeleteUser(r.Context(), uint(id)); err != nil {
		http.Error(w, "Could not delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ✅ Fix: Ensure getCoursesHandler is inside application struct
func (app *application) getCoursesHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Get Courses Handler"}
	json.NewEncoder(w).Encode(response)
}

// ✅ Fix: Ensure createCourseHandler is inside application struct
func (app *application) createCourseHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Create Course Handler"}
	json.NewEncoder(w).Encode(response)
}

// ✅ Fix: Ensure getInstructorsHandler is inside application struct
func (app *application) getInstructorsHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Get Instructors Handler"}
	json.NewEncoder(w).Encode(response)
}

// ✅ Fix: Ensure createInstructorHandler is inside application struct
func (app *application) createInstructorHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Create Instructor Handler"}
	json.NewEncoder(w).Encode(response)
}

// ✅ Fix: Ensure getTraceHandler is inside application struct
func (app *application) getTraceHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Get Trace Handler"}
	json.NewEncoder(w).Encode(response)
}

// ✅ Fix: Ensure uploadTraceHandler is inside application struct
func (app *application) uploadTraceHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Upload Trace Handler"}
	json.NewEncoder(w).Encode(response)
}

package main

import (
	"github.com/csye7125/team01/internal/handlers"
	"github.com/csye7125/team01/internal/middlewares"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
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

	healthHandler := handlers.NewHealthHandler()
	userHandler := handlers.NewUserHandler(a.store)
	courseHandler := handlers.NewCourseHandler(a.store)
	instructorHandler := handlers.NewInstructorHandler(a.store)
	traceHandler := handlers.NewTraceHandler(a.store, os.Getenv("GCS_BUCKET_NAME"))
	authMiddleware := middlewares.NewAuthMiddleware(a.store.Users)

	r.Get("/healthz", healthHandler.HealthCheckHandler)
	r.Post("/v1/user", userHandler.CreateUserHandler)
	r.Get("/v1/course/{courseId}", courseHandler.GetCourseHandler)
	r.Get("/v1/instructor/{instructorId}", instructorHandler.GetInstructorHandler)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.BasicAuthMiddleware)

		r.Get("/v1/user/{userId}", userHandler.GetUserHandler)
		r.Put("/v1/user/{userId}", userHandler.UpdateUserHandler)
		r.Delete("/v1/user/{userId}", userHandler.DeleteUserHandler)

		r.Post("/v1/course", courseHandler.CreateCourseHandler)
		r.Put("/v1/course/{courseId}", courseHandler.UpdateCourseHandler)
		r.Patch("/v1/course/{courseId}", courseHandler.PatchCourseHandler)
		r.Delete("/v1/course/{courseId}", courseHandler.DeleteCourseHandler)

		r.Post("/v1/instructor", instructorHandler.CreateInstructorHandler)
		r.Put("/v1/instructor/{instructorId}", instructorHandler.UpdateInstructorHandler)
		r.Patch("/v1/instructor/{instructorId}", instructorHandler.PatchInstructorHandler)
		r.Delete("/v1/instructor/{instructorId}", instructorHandler.DeleteInstructorHandler)

		r.Post("/v1/course/{course_id}/trace", traceHandler.UploadTraceHandler)
		r.Get("/v1/course/{course_id}/trace/{trace_id}", traceHandler.GetTraceHandler)
		r.Get("/v1/course/{course_id}/trace", traceHandler.GetAllTracesHandler)
		r.Delete("/v1/course/{course_id}/trace/{trace_id}", traceHandler.DeleteTraceHandler)
	})
	return r
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

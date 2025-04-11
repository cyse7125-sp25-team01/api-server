package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/csye7125/team01/internal/handlers"
	"github.com/csye7125/team01/internal/middlewares"
	"github.com/csye7125/team01/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

// Wrap a handler with OpenTelemetry instrumentation
func wrapHandler(handler http.HandlerFunc, operation string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		otelhttp.NewHandler(http.HandlerFunc(handler), operation).ServeHTTP(w, r)
	}
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

	// Public endpoints with OpenTelemetry instrumentation
	r.Get("/healthz", wrapHandler(healthHandler.HealthCheckHandler, "HealthCheck"))
	r.Post("/v1/user", wrapHandler(userHandler.CreateUserHandler, "CreateUser"))
	r.Get("/v1/course/{courseId}", wrapHandler(courseHandler.GetCourseHandler, "GetCourse"))
	r.Get("/v1/instructor/{instructorId}", wrapHandler(instructorHandler.GetInstructorHandler, "GetInstructor"))

	// Protected endpoints with OpenTelemetry instrumentation
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.BasicAuthMiddleware)

		r.Get("/v1/user/{userId}", wrapHandler(userHandler.GetUserHandler, "GetUser"))
		r.Put("/v1/user/{userId}", wrapHandler(userHandler.UpdateUserHandler, "UpdateUser"))
		r.Delete("/v1/user/{userId}", wrapHandler(userHandler.DeleteUserHandler, "DeleteUser"))

		r.Post("/v1/course", wrapHandler(courseHandler.CreateCourseHandler, "CreateCourse"))
		r.Put("/v1/course/{courseId}", wrapHandler(courseHandler.UpdateCourseHandler, "UpdateCourse"))
		r.Patch("/v1/course/{courseId}", wrapHandler(courseHandler.PatchCourseHandler, "PatchCourse"))
		r.Delete("/v1/course/{courseId}", wrapHandler(courseHandler.DeleteCourseHandler, "DeleteCourse"))

		r.Post("/v1/instructor", wrapHandler(instructorHandler.CreateInstructorHandler, "CreateInstructor"))
		r.Put("/v1/instructor/{instructorId}", wrapHandler(instructorHandler.UpdateInstructorHandler, "UpdateInstructor"))
		r.Patch("/v1/instructor/{instructorId}", wrapHandler(instructorHandler.PatchInstructorHandler, "PatchInstructor"))
		r.Delete("/v1/instructor/{instructorId}", wrapHandler(instructorHandler.DeleteInstructorHandler, "DeleteInstructor"))

		r.Post("/v1/course/{course_id}/trace", wrapHandler(traceHandler.UploadTraceHandler, "UploadTrace"))
		r.Get("/v1/course/{course_id}/trace/{trace_id}", wrapHandler(traceHandler.GetTraceHandler, "GetTrace"))
		r.Get("/v1/course/{course_id}/trace", wrapHandler(traceHandler.GetAllTracesHandler, "GetAllTraces"))
		r.Delete("/v1/course/{course_id}/trace/{trace_id}", wrapHandler(traceHandler.DeleteTraceHandler, "DeleteTrace"))
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
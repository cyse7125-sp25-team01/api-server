package store

import (
	"gorm.io/gorm"
)

// Storage struct holds all database repositories
type Storage struct {
	DB          *gorm.DB
	Users       *UserStore
	Traces      *TraceStore
	Courses     *CourseStore
	Instructors *InstructorStore
}

// NewStorage initializes Storage with a database connection
func NewStorage(db *gorm.DB) *Storage {
	return &Storage{
		DB:          db,
		Users:       NewUserStore(db),
		Traces:      NewTraceStore(db),
		Courses:     NewCourseStore(db),
		Instructors: NewInstructorStore(db),
	}
}

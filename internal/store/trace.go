package store

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Trace struct {
	TraceID     uint      `json:"trace_id" gorm:"primaryKey;autoIncrement"`
	CourseID    uint      `json:"course_id"`
	UserID      uint      `json:"user_id"`
	FileName    string    `json:"file_name"`
	DateCreated time.Time `json:"date_created"`
	BucketPath  string    `json:"bucket_path"`
}

type TraceStore struct {
	db *gorm.DB
}

func NewTraceStore(db *gorm.DB) *TraceStore {
	return &TraceStore{db: db}
}

// Create Trace
func (s *TraceStore) CreateTrace(ctx context.Context, trace *Trace) error {
	return s.db.WithContext(ctx).Create(&trace).Error
}

// Get Trace by ID
func (s *TraceStore) GetTraceByID(ctx context.Context, courseID, traceID string) (*Trace, error) {
	var trace Trace
	err := s.db.WithContext(ctx).Where("course_id = ? AND trace_id = ?", courseID, traceID).First(&trace).Error
	if err != nil {
		return nil, err
	}
	return &trace, nil
}

// Get All Traces by Course ID
func (s *TraceStore) GetTracesByCourseID(ctx context.Context, courseID string) ([]Trace, error) {
	var traces []Trace
	err := s.db.WithContext(ctx).Where("course_id = ?", courseID).Find(&traces).Error
	if err != nil {
		return nil, err
	}
	return traces, nil
}

// Delete Trace
func (s *TraceStore) DeleteTrace(ctx context.Context, courseID, traceID string) error {
	return s.db.WithContext(ctx).Where("course_id = ? AND trace_id = ?", courseID, traceID).Delete(&Trace{}).Error
}

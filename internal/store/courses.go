package store

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type Course struct {
	ID              uint      `gorm:"primaryKey;column:course_id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	SemesterTerm    string    `json:"semester_term"`
	Manufacturer    string    `json:"manufacturer"`
	CreditHours     int       `json:"credit_hours"`
	SemesterYear    int       `json:"semester_year"`
	DateAdded       time.Time `json:"date_added"`
	DateLastUpdated time.Time `json:"date_last_updated"`
	OwnerUserID     uint      `json:"owner_user_id"`
	InstructorID    uint      `json:"instructor_id"`
}

type CourseStore struct {
	db *gorm.DB
}

func NewCourseStore(db *gorm.DB) *CourseStore {
	return &CourseStore{db: db}
}

func (s *CourseStore) CreateCourse(ctx context.Context, course *Course) error {
	return s.db.WithContext(ctx).Create(&course).Error
}

func (s *CourseStore) GetCourseByID(ctx context.Context, id uint) (*Course, error) {
	var course Course
	if err := s.db.WithContext(ctx).First(&course, id).Error; err != nil {
		return nil, err
	}
	return &course, nil
}

// UpdateCourse updates an existing course
func (s *CourseStore) UpdateCourse(ctx context.Context, id uint, updateData *Course) error {
	return s.db.WithContext(ctx).Model(&Course{}).Where("course_id = ?", id).Updates(updateData).Error
}

// PatchCourse performs a partial update on a course
func (s *CourseStore) PatchCourse(ctx context.Context, id uint, updateData map[string]interface{}) error {
	return s.db.WithContext(ctx).Model(&Course{}).Where("course_id = ?", id).Updates(updateData).Error
}

// DeleteCourse removes a course from the database
func (s *CourseStore) DeleteCourse(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&Course{}, id).Error
}

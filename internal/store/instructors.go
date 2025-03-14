package store

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Instructor struct {
	InstructorID uint      `json:"instructor_id" gorm:"primaryKey;autoIncrement"`
	UserID       uint      `json:"user_id"`
	Name         string    `json:"name"`
	DateCreated  time.Time `json:"date_created" gorm:"default:CURRENT_TIMESTAMP"`
}

type InstructorStore struct {
	db *gorm.DB
}

func NewInstructorStore(db *gorm.DB) *InstructorStore {
	return &InstructorStore{db: db}
}

func (s *InstructorStore) CreateInstructor(ctx context.Context, username string, instructor *Instructor) error {
	// 🔹 Step 1: Get User ID using username
	var user User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user with username '%s' does not exist", username)
	}

	// 🔹 Step 2: Assign the retrieved user ID
	instructor.UserID = user.ID

	// 🔹 Step 3: Insert the instructor into the database
	return s.db.WithContext(ctx).Create(&instructor).Error
}

func (s *InstructorStore) GetInstructorByID(ctx context.Context, id string) (*Instructor, error) {
	var instructor Instructor
	if err := s.db.WithContext(ctx).First(&instructor, "instructor_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &instructor, nil
}

func (s *InstructorStore) UpdateInstructor(ctx context.Context, id string, updateData *Instructor) error {
	return s.db.WithContext(ctx).Model(&Instructor{}).Where("instructor_id = ?", id).Updates(updateData).Error
}

func (s *InstructorStore) DeleteInstructor(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&Instructor{}, "instructor_id = ?", id).Error
}

func (s *InstructorStore) CheckInstructorExists(ctx context.Context, instructorID uint) error {
	var instructor Instructor
	if err := s.db.WithContext(ctx).First(&instructor, "instructor_id = ?", instructorID).Error; err != nil {
		return fmt.Errorf("instructor not found")
	}
	return nil
}

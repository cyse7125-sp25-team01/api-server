package store

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// User struct represents the user model based on Swagger API Docs
type User struct {
	ID             uint      `gorm:"primaryKey" json:"id,omitempty"`
	FirstName      string    `gorm:"not null" json:"first_name" validate:"required"`
	LastName       string    `gorm:"not null" json:"last_name" validate:"required"`
	Password       string    `gorm:"not null" json:"password,omitempty" validate:"required"` // Write-only
	Username       string    `gorm:"unique;not null" json:"username" validate:"required,email"`
	AccountCreated time.Time `gorm:"autoCreateTime" json:"account_created,omitempty"`
	AccountUpdated time.Time `gorm:"autoUpdateTime" json:"account_updated,omitempty"`
}

// UserStore represents the user repository using GORM
type UserStore struct {
	db *gorm.DB
}

func (s *UserStore) UpdateUser(ctx context.Context, id uint, updateData *User) error {
	return s.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updateData).Error
}

func (s *UserStore) DeleteUser(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&User{}, id).Error
}

// ✅ NewUserStore initializes the UserStore with *gorm.DB
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

// ✅ HashPassword hashes the password before storing it
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ✅ CreateUser inserts a new user into the database using GORM
func (s *UserStore) CreateUser(ctx context.Context, user *User) error {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword // Store hashed password
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return err
	}
	return nil
}

// ✅ GetUserByID fetches a user by ID using GORM
func (s *UserStore) GetUserByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

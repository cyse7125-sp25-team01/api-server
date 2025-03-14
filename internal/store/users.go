package store

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

// ✅ User Model
type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Password       string    `json:"password"`
	Username       string    `json:"username" gorm:"unique"`
	AccountCreated time.Time `json:"account_created" gorm:"autoCreateTime"`
	AccountUpdated time.Time `json:"account_updated" gorm:"autoUpdateTime"`
}

// ✅ UserStore Struct
type UserStore struct {
	db *gorm.DB
}

// ✅ NewUserStore Initializes UserStore with *gorm.DB
func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

// ✅ HashPassword to Securely Hash User Passwords
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ✅ CreateUser - Insert New User Into DB
func (s *UserStore) CreateUser(ctx context.Context, user *User) error {
	// Check if the user already exists
	var existingUser User
	if err := s.db.WithContext(ctx).Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return gorm.ErrDuplicatedKey // Return a duplicate key error
	}

	// Hash password before storing
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// Insert user into DB
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return err
	}

	// Exclude password from response
	user.Password = ""
	return nil
}

// ✅ GetUserByID - Retrieve User by ID
func (s *UserStore) GetUserByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	// Exclude password from response
	user.Password = ""
	return &user, nil
}

// ✅ UpdateUser - Modify User Info (Handles Password Hashing)
func (s *UserStore) UpdateUser(ctx context.Context, id uint, updateData *User) error {
	// Hash password if it's being updated
	if updateData.Password != "" {
		hashedPassword, err := HashPassword(updateData.Password)
		if err != nil {
			return err
		}
		updateData.Password = hashedPassword
	}

	// Update user in DB
	return s.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Updates(updateData).Error
}

// ✅ DeleteUser - Remove User From DB (Ensures User Exists First)
func (s *UserStore) DeleteUser(ctx context.Context, id uint) error {
	// Check if the user exists before deletion
	var user User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return err // User not found
	}

	// Proceed with deletion
	return s.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (s *UserStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetUserByCredentials(ctx context.Context, username, password string) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Compare Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("incorrect password")
	}

	return &user, nil
}

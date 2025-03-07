package store

import "gorm.io/gorm"

type Storage struct {
	Users *UserStore
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{
		Users: &UserStore{db},
	}
}

package repositories

import "errors"

var (
	ErrDuplicateUsername = errors.New("username sudah digunakan")
	ErrNotFound          = errors.New("data tidak ditemukan")
)
// Package repository provides SQLite repository implementations.
package repository

import "errors"

// ErrNotFound is returned when an item is not found in the database.
var ErrNotFound = errors.New("not found")

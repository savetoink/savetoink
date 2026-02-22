package repository

import "errors"

// ErrNotFound is returned when an article is not found.
var ErrNotFound = errors.New("article not found")

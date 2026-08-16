package store

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("conflict")
	ErrNoHealthEvents = errors.New("no health events")
)

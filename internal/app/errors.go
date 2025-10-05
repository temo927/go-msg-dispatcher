package app

import "errors"

var (
	ErrSendFailed     = errors.New("provider send failed")
	ErrMarkSent       = errors.New("mark sent failed")
	ErrMarkFailed     = errors.New("mark failed")
	ErrAlreadyRunning = errors.New("scheduler already running")
	ErrNotRunning     = errors.New("scheduler not running")
)

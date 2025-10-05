package app

import "errors"

var (
	ErrAlreadyRunning = errors.New("scheduler already running")
	ErrNotRunning     = errors.New("scheduler not running")
	ErrSendFailed     = errors.New("failed to send message")
	ErrNoMessages     = errors.New("no queued messages to process")
)

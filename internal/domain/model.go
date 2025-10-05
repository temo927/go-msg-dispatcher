package domain

import "time"

type Message struct {
	ID                 string
	ToPhone            string
	Content            string
	Status             string
	RetryCount         int
	ProviderMessageID  *string
	LastError          *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	SentAt             *time.Time
}

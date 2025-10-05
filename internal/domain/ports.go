package domain

import "context"

type MessagesRepo interface {
	ClaimNextBatch(ctx context.Context, limit int) ([]Message, error)
	MarkSent(ctx context.Context, id, providerMessageID string) error
	MarkFailed(ctx context.Context, id string, err error, maxRetries int) error
	ListSent(ctx context.Context, limit, offset int) ([]Message, error)
	Create(ctx context.Context, to, content string) (Message, error)
}

type Provider interface {
	Send(ctx context.Context, msg Message) (string, error)
}

type Cache interface {
	SetSentMeta(ctx context.Context, msgID string, meta map[string]string) error
}

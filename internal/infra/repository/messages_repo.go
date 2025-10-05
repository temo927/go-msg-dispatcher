package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/temo927/go-msg-dispatcher/internal/domain"
)

type MessagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) *MessagesRepo {
	return &MessagesRepo{db: db}
}

func (r *MessagesRepo) ClaimNextBatch(ctx context.Context, limit int) ([]domain.Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
    UPDATE messages
    SET status = 'processing'::message_status, updated_at = NOW()
    WHERE id IN (
        SELECT id FROM messages
        WHERE status = 'queued'::message_status
        ORDER BY created_at
        FOR UPDATE SKIP LOCKED
        LIMIT $1
    )
    RETURNING id, to_phone, content, status, retry_count, provider_message_id, last_error, created_at, updated_at, sent_at
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(
			&m.ID, &m.ToPhone, &m.Content, &m.Status, &m.RetryCount,
			&m.ProviderMessageID, &m.LastError, &m.CreatedAt, &m.UpdatedAt, &m.SentAt,
		); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *MessagesRepo) MarkSent(ctx context.Context, id, providerMessageID string) error {
	_, err := r.db.ExecContext(ctx, `
    UPDATE messages
    SET status = 'sent'::message_status, provider_message_id = $2, sent_at = NOW(), updated_at = NOW()
    WHERE id = $1
`, id, providerMessageID)
	return err
}

func (r *MessagesRepo) MarkFailed(ctx context.Context, id string, cause error, maxRetries int) error {
	_, err := r.db.ExecContext(ctx, `
    UPDATE messages
    SET status = CASE WHEN retry_count + 1 >= $2 THEN 'failed'::message_status ELSE 'queued'::message_status END,
        retry_count = retry_count + 1,
        last_error = $3,
        updated_at = NOW()
    WHERE id = $1
`, id, maxRetries, cause.Error())
	return err
}

func (r *MessagesRepo) ListSent(ctx context.Context, limit, offset int) ([]domain.Message, error) {
	rows, err := r.db.QueryContext(ctx, `
    SELECT id, to_phone, content, status, retry_count, provider_message_id, last_error, created_at, updated_at, sent_at
    FROM messages
    WHERE status = 'sent'::message_status
    ORDER BY sent_at DESC
    LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(
			&m.ID, &m.ToPhone, &m.Content, &m.Status, &m.RetryCount,
			&m.ProviderMessageID, &m.LastError, &m.CreatedAt, &m.UpdatedAt, &m.SentAt,
		); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *MessagesRepo) Create(ctx context.Context, to, content string) (domain.Message, error) {
	var m domain.Message
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO messages (to_phone, content)
		VALUES ($1, $2)
		RETURNING id, to_phone, content, status, retry_count, provider_message_id, last_error, created_at, updated_at, sent_at
	`, to, content).Scan(
		&m.ID, &m.ToPhone, &m.Content, &m.Status, &m.RetryCount,
		&m.ProviderMessageID, &m.LastError, &m.CreatedAt, &m.UpdatedAt, &m.SentAt,
	)
	if err != nil {
		return domain.Message{}, err
	}
	return m, nil
}

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

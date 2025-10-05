package app

import (
	"context"
	"fmt"
	"time"

	"github.com/temo927/go-msg-dispatcher/internal/domain"
	"github.com/temo927/go-msg-dispatcher/internal/infra/log"
)

type Sender struct {
	repo  domain.MessagesRepo
	prov  domain.Provider
	cache domain.Cache
	cfg   SenderConfig
}

type SenderConfig struct {
	MaxRetries int
}

func NewSender(repo domain.MessagesRepo, prov domain.Provider, cache domain.Cache, cfg SenderConfig) *Sender {
	return &Sender{repo: repo, prov: prov, cache: cache, cfg: cfg}
}

func (s *Sender) Send(ctx context.Context, msg domain.Message) error {
	providerID, err := s.prov.Send(ctx, msg)
	if err != nil {
		if e := s.repo.MarkFailed(ctx, msg.ID, err, s.cfg.MaxRetries); e != nil {
			log.Logger.Error("mark failed", "msg_id", msg.ID, "err", e)
			return fmt.Errorf("mark failed: %w", e)
		}
		log.Logger.Error("provider send failed", "msg_id", msg.ID, "err", err)
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	if err := s.repo.MarkSent(ctx, msg.ID, providerID); err != nil {
		log.Logger.Error("mark sent failed", "msg_id", msg.ID, "err", err)
		return fmt.Errorf("mark sent: %w", err)
	}

	if s.cache != nil {
		_ = s.cache.SetSentMeta(ctx, msg.ID, map[string]string{
			"messageId": providerID,
			"sent_at":   time.Now().UTC().Format(time.RFC3339),
		})
	}

	log.Logger.Info("provider send ok", "msg_id", msg.ID, "provider_message_id", providerID)
	return nil
}

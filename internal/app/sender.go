// internal/app/sender.go
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/temo927/go-msg-dispatcher/internal/domain"
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
	return &Sender{
		repo:  repo,
		prov:  prov,
		cache: cache,
		cfg:   cfg,
	}
}

func (s *Sender) Send(ctx context.Context, msg domain.Message) error {
	providerID, err := s.prov.Send(ctx, msg)
	if err != nil {
	
		if e := s.repo.MarkFailed(ctx, msg.ID, err, s.cfg.MaxRetries); e != nil {
			return fmt.Errorf("provider send failed: %v (mark failed error: %v)", err, e)
		}
		return fmt.Errorf("provider send failed: %v", err)
	}

	if err := s.repo.MarkSent(ctx, msg.ID, providerID); err != nil {
		return fmt.Errorf("mark sent failed: %v", err)
	}

	if s.cache != nil {
		_ = s.cache.SetSentMeta(ctx, msg.ID, map[string]string{
			"messageId": providerID,
			"sent_at":   time.Now().UTC().Format(time.RFC3339),
		})
	}

	return nil
}

package app

import (
	"context"
	"time"

	"github.com/temo927/go-msg-dispatcher/internal/domain"
	"github.com/temo927/go-msg-dispatcher/internal/infra/log"
)

type Scheduler struct {
	repo      domain.MessagesRepo
	sender    *Sender
	ticker    *time.Ticker
	cancel    context.CancelFunc
	running   bool
	interval  time.Duration
	batchSize int
}

func NewScheduler(repo domain.MessagesRepo, sender *Sender, interval time.Duration, batchSize int) *Scheduler {
	return &Scheduler{repo: repo, sender: sender, interval: interval, batchSize: batchSize}
}

func (s *Scheduler) Start(ctx context.Context) error {
	if s.running {
		return ErrAlreadyRunning
	}
	ctx, s.cancel = context.WithCancel(ctx)
	s.ticker = time.NewTicker(s.interval)
	s.running = true
	log.Logger.Info("scheduler started", "interval", s.interval, "batch_size", s.batchSize)
	go s.loop(ctx)
	return nil
}

func (s *Scheduler) Stop() error {
	if !s.running {
		return ErrNotRunning
	}
	s.ticker.Stop()
	s.cancel()
	s.running = false
	log.Logger.Info("scheduler stopped")
	return nil
}

func (s *Scheduler) IsRunning() bool { return s.running }

func (s *Scheduler) loop(ctx context.Context) {
	for {
		select {
		case <-s.ticker.C:
			if err := s.process(ctx); err != nil {
				log.Logger.Error("scheduler tick failed", "err", err)
			}
		case <-ctx.Done():
			log.Logger.Info("scheduler exiting")
			return
		}
	}
}

func (s *Scheduler) process(ctx context.Context) error {
	msgs, err := s.repo.ClaimNextBatch(ctx, s.batchSize)
	if err != nil {
		log.Logger.Error("claim batch failed", "err", err)
		return err
	}
	if len(msgs) == 0 {
		log.Logger.Info("no queued messages to process")
		return ErrNoMessages
	}
	for _, m := range msgs {
		if err := s.sender.Send(ctx, m); err != nil {
			log.Logger.Error("send failed", "msg_id", m.ID, "err", err)
		} else {
			log.Logger.Info("message sent", "msg_id", m.ID)
		}
	}
	return nil
}

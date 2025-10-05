package app

import (
	"context"
	"sync"
	"time"

	"github.com/temo927/go-msg-dispatcher/internal/domain"
	"github.com/temo927/go-msg-dispatcher/internal/infra/log"
)

type Scheduler struct {
	repo      domain.MessagesRepo
	sender    *Sender
	interval  time.Duration
	batchSize int

	mu     sync.Mutex
	ticker *time.Ticker
	cancel context.CancelFunc
	running bool
}

func NewScheduler(repo domain.MessagesRepo, sender *Sender, interval time.Duration, batchSize int) *Scheduler {
	return &Scheduler{
		repo: repo, sender: sender,
		interval: interval, batchSize: batchSize,
	}
}

func (s *Scheduler) Start(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		log.Logger.Info("scheduler already running")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.ticker = time.NewTicker(s.interval)
	s.running = true

	log.Logger.Info("scheduler started", "interval", s.interval, "batch_size", s.batchSize)

	go s.loop(ctx)
	return nil
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		log.Logger.Info("scheduler already stopped")
		return nil
	}

	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.cancel != nil {
		s.cancel()
	}
	s.ticker = nil
	s.cancel = nil
	s.running = false

	log.Logger.Info("scheduler stopped")
	return nil
}

func (s *Scheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) loop(ctx context.Context) {
	if err := s.process(ctx); err != nil {
		log.Logger.Error("scheduler initial process failed", "err", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Logger.Info("scheduler exiting")
			return
		case <-s.ticker.C:
			if err := s.process(ctx); err != nil {
				log.Logger.Error("scheduler tick failed", "err", err)
			}
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
		return nil
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

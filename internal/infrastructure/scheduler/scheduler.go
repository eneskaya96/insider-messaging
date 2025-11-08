package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eneskaya/insider-messaging/internal/application/service"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"go.uber.org/zap"
)

type Scheduler struct {
	messageService service.MessageService
	batchSize      int
	interval       time.Duration
	workerCount    int

	mu           sync.RWMutex
	isRunning    bool
	stopChan     chan struct{}
	stoppedChan  chan struct{}
	wg           sync.WaitGroup

	lastRunAt       time.Time
	totalProcessed  int64
	totalSuccessful int64
	totalFailed     int64
}

func NewScheduler(
	messageService service.MessageService,
	batchSize int,
	intervalSeconds int,
	workerCount int,
) *Scheduler {
	return &Scheduler{
		messageService: messageService,
		batchSize:      batchSize,
		interval:       time.Duration(intervalSeconds) * time.Second,
		workerCount:    workerCount,
		stopChan:       make(chan struct{}),
		stoppedChan:    make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		logger.Get().Warn("scheduler is already running")
		return nil
	}
	s.isRunning = true
	s.stopChan = make(chan struct{})
	s.stoppedChan = make(chan struct{})
	s.mu.Unlock()

	logger.Get().Info("starting message scheduler",
		zap.Int("batch_size", s.batchSize),
		zap.Duration("interval", s.interval),
		zap.Int("worker_count", s.workerCount),
	)

	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		logger.Get().Warn("scheduler is not running")
		return nil
	}
	s.mu.Unlock()

	logger.Get().Info("stopping message scheduler")

	close(s.stopChan)

	s.wg.Wait()

	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	close(s.stoppedChan)

	logger.Get().Info("message scheduler stopped successfully")
	return nil
}

func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *Scheduler) GetStats() (lastRunAt time.Time, processed, successful, failed int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRunAt, atomic.LoadInt64(&s.totalProcessed), atomic.LoadInt64(&s.totalSuccessful), atomic.LoadInt64(&s.totalFailed)
}

func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.processMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Get().Info("scheduler context cancelled")
			return
		case <-s.stopChan:
			logger.Get().Info("scheduler stop signal received")
			return
		case <-ticker.C:
			s.processMessages(ctx)
		}
	}
}

func (s *Scheduler) processMessages(ctx context.Context) {
	s.mu.Lock()
	s.lastRunAt = time.Now()
	s.mu.Unlock()

	logger.Get().Info("starting message processing cycle")

	processCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	jobsChan := make(chan struct{}, s.batchSize)
	resultsChan := make(chan bool, s.batchSize)

	var workerWg sync.WaitGroup
	for i := 0; i < s.workerCount; i++ {
		workerWg.Add(1)
		go s.worker(processCtx, i, jobsChan, resultsChan, &workerWg)
	}

	go func() {
		for i := 0; i < s.batchSize; i++ {
			select {
			case <-processCtx.Done():
				return
			case jobsChan <- struct{}{}:
			}
		}
		close(jobsChan)
	}()

	go func() {
		workerWg.Wait()
		close(resultsChan)
	}()

	successful := int64(0)
	failed := int64(0)
	for result := range resultsChan {
		if result {
			successful++
		} else {
			failed++
		}
	}

	processed := successful + failed
	atomic.AddInt64(&s.totalProcessed, processed)
	atomic.AddInt64(&s.totalSuccessful, successful)
	atomic.AddInt64(&s.totalFailed, failed)

	logger.Get().Info("message processing cycle completed",
		zap.Int64("processed", processed),
		zap.Int64("successful", successful),
		zap.Int64("failed", failed),
	)
}

func (s *Scheduler) worker(ctx context.Context, id int, jobs <-chan struct{}, results chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-jobs:
			if !ok {
				return
			}

			_, err := s.messageService.ProcessPendingMessages(ctx, 1)
			results <- (err == nil)
		}
	}
}

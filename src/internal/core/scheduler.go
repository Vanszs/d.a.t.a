package core

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	jobs     map[string]*Job
	mu       sync.RWMutex
	shutdown chan struct{}
}

type Job struct {
	ID       string
	Interval time.Duration
	LastRun  time.Time
	JobFunc  func() error
	ticker   *time.Ticker
	cancel   context.CancelFunc
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:     make(map[string]*Job),
		shutdown: make(chan struct{}),
	}
}

func (s *Scheduler) SchedulePeriodic(ctx context.Context, interval time.Duration, jobFunc func() error, wg *sync.WaitGroup) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobCtx, cancel := context.WithCancel(ctx)
	job := &Job{
		ID:       uuid.New().String(),
		Interval: interval,
		JobFunc:  jobFunc,
		ticker:   time.NewTicker(interval),
		cancel:   cancel,
	}

	go func() {
		defer wg.Done()
		for {
			select {
			case <-job.ticker.C:
				// TODO: add error handling
				job.JobFunc()
				job.LastRun = time.Now()
			case <-jobCtx.Done():
				job.ticker.Stop()
				return
			case <-s.shutdown:
				job.ticker.Stop()
				return
			}
		}
	}()

	s.jobs[job.ID] = job
	return job.ID
}

func (s *Scheduler) CancelJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, exists := s.jobs[id]; exists {
		job.cancel()
		delete(s.jobs, id)
	}
}

func (s *Scheduler) Shutdown() {
	close(s.shutdown)
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range s.jobs {
		job.cancel()
	}
	s.jobs = make(map[string]*Job)
}

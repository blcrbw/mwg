package processor

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// LockProvider defines a distributed Lock provider.
type LockProvider interface {
	Acquire(ctx context.Context, key string) (bool, func() error, error)
}

// Processor defines a movie processor.
type Processor struct {
	logger       *zap.Logger
	lockProvider LockProvider
}

// New creates a new movie processor.
func New(logger *zap.Logger, lockProvider LockProvider) *Processor {
	return &Processor{
		logger:       logger,
		lockProvider: lockProvider,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	p.logger.Info("Starting the movie processor")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, release, err := p.lockProvider.Acquire(ctx, "locks/service/movie/processor")
			if err != nil {
				p.logger.Error("Unable to acquire lock, retrying in 10 seconds", zap.Error(err))
				time.Sleep(10 * time.Second)
				continue
			}
			p.logger.Info("Lock has been acquired, starting processing")
			if err := p.process(ctx); err != nil {
				p.logger.Error("Process error", zap.Error(err))
			} else {
				p.logger.Info("Process completed successfully")
			}
			p.logger.Info("Releasing the lock")
			if err := release(); err != nil {
				p.logger.Error("Failed to release the lock", zap.Error(err))
			}
		}
	}
}

// process executes processing logic.
func (p *Processor) process(ctx context.Context) error {
	const timeout = 5 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Dummy simulation.
	time.Sleep(5 * time.Second)
	return nil
}

package limiter

import (
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Limiter struct {
	logger *zap.Logger
	l      *rate.Limiter
}

func New(logger *zap.Logger, limit, burst int) *Limiter {
	return &Limiter{logger: logger, l: rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *Limiter) Limit() bool {
	allowed := l.l.Allow()
	l.logger.Debug("Rate limit check",
		zap.Bool("allowed", allowed),
		zap.Float64("limit", float64(l.l.Limit())),
		zap.Int("burst", l.l.Burst()),
	)
	return !allowed
}

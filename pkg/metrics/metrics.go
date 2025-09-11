package metrics

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/uber-go/tally/v6"
	"github.com/uber-go/tally/v6/prometheus"
	"go.uber.org/zap"
)

func NewMetricsReporter(logger *zap.Logger, serviceName string, metricsPort int) (scope tally.Scope, closer io.Closer) {
	reporter := prometheus.NewReporter(prometheus.Options{})
	scope, closer = tally.NewRootScope(tally.ScopeOptions{
		Tags:            map[string]string{"service": serviceName},
		CachedReporter:  reporter,
		SanitizeOptions: &prometheus.DefaultSanitizerOpts,
	}, 10*time.Second)
	http.Handle("/metrics", reporter.HTTPHandler())
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			logger.Fatal("Failed to start metrics handler", zap.Error(err))
		}
	}()

	counter := scope.Counter("service_started")
	counter.Inc(1)
	return scope, closer
}

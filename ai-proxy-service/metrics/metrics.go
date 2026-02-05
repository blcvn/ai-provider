package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal tracks total completion requests
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_proxy_requests_total",
			Help: "Total number of completion requests",
		},
		[]string{"model_id", "provider", "status"},
	)

	// RequestDuration tracks request latency
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_proxy_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model_id", "provider"},
	)

	// TokensUsed tracks token consumption
	TokensUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_proxy_tokens_used_total",
			Help: "Total tokens used",
		},
		[]string{"model_id", "provider", "type"}, // type: prompt, completion
	)

	// CostTotal tracks cumulative cost
	CostTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_proxy_cost_total",
			Help: "Total cost in USD",
		},
		[]string{"model_id", "provider"},
	)

	// CacheHits tracks cache hit/miss
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_proxy_cache_hits_total",
			Help: "Total cache hits",
		},
		[]string{"status"}, // status: hit, miss
	)

	// CircuitBreakerState tracks circuit breaker states
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ai_proxy_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"provider"},
	)
)

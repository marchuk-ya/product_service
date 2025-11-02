package metrics

import (
	"product_service/products/internal/usecase/ports"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var _ ports.MetricsCollector = (*prometheusMetrics)(nil)

type prometheusMetrics struct {
	productsCreatedTotal      prometheus.Counter
	productsDeletedTotal      prometheus.Counter
	requestDuration           *prometheus.HistogramVec
	requestCount              *prometheus.CounterVec
	databaseQueryDuration     prometheus.Histogram
	rabbitmqPublishDuration   prometheus.Histogram
	transactionsRetryTotal     prometheus.Counter
	transactionsRetrySuccess   prometheus.Counter
	transactionsRetryFailed    prometheus.Counter
	batchSize                  *prometheus.HistogramVec
	outboxRetryAttempts        *prometheus.HistogramVec
	outboxEventsProcessed       *prometheus.CounterVec
}

func NewPrometheusMetrics() ports.MetricsCollector {
	return &prometheusMetrics{
		productsCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "products_created_total",
			Help: "Total number of products created",
		}),
		productsDeletedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "products_deleted_total",
			Help: "Total number of products deleted",
		}),
		requestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "endpoint", "status"}),
		requestCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "endpoint", "status"}),
		databaseQueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		}),
		rabbitmqPublishDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "rabbitmq_publish_duration_seconds",
			Help:    "RabbitMQ publish duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		}),
		transactionsRetryTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "transactions_retry_total",
			Help: "Total number of transaction retries",
		}),
		transactionsRetrySuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "transactions_retry_success_total",
			Help: "Total number of successful transactions after retry",
		}),
		transactionsRetryFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "transactions_retry_failed_total",
			Help: "Total number of failed transactions after retry",
		}),
		batchSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "batch_operation_size",
			Help:    "Size of batch operations",
			Buckets: []float64{1, 5, 10, 25, 50, 100},
		}, []string{"operation"}),
		outboxRetryAttempts: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "outbox_retry_attempts",
			Help:    "Number of retry attempts for outbox events",
			Buckets: []float64{1, 2, 3, 4, 5, 10},
		}, []string{"event_type"}),
		outboxEventsProcessed: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "outbox_events_processed_total",
			Help: "Total number of outbox events processed",
		}, []string{"event_type", "status"}),
	}
}

func (m *prometheusMetrics) IncrementProductsCreated() {
	m.productsCreatedTotal.Inc()
}

func (m *prometheusMetrics) IncrementProductsDeleted() {
	m.productsDeletedTotal.Inc()
}

func (m *prometheusMetrics) RecordRequestDuration(method, endpoint, status string, duration time.Duration) {
	m.requestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
}

func (m *prometheusMetrics) IncrementRequestCount(method, endpoint, status string) {
	m.requestCount.WithLabelValues(method, endpoint, status).Inc()
}

func (m *prometheusMetrics) RecordDatabaseQueryDuration(duration time.Duration) {
	m.databaseQueryDuration.Observe(duration.Seconds())
}

func (m *prometheusMetrics) RecordRabbitMQPublishDuration(duration time.Duration) {
	m.rabbitmqPublishDuration.Observe(duration.Seconds())
}

func (m *prometheusMetrics) IncrementTransactionRetry() {
	m.transactionsRetryTotal.Inc()
}

func (m *prometheusMetrics) IncrementTransactionRetrySuccess() {
	m.transactionsRetrySuccess.Inc()
}

func (m *prometheusMetrics) IncrementTransactionRetryFailed() {
	m.transactionsRetryFailed.Inc()
}

func (m *prometheusMetrics) RecordBatchSize(operation string, size int) {
	m.batchSize.WithLabelValues(operation).Observe(float64(size))
}

func (m *prometheusMetrics) RecordOutboxRetryAttempt(eventType string, attempt int) {
	m.outboxRetryAttempts.WithLabelValues(eventType).Observe(float64(attempt))
}

func (m *prometheusMetrics) RecordOutboxEventProcessed(eventType string, status string) {
	m.outboxEventsProcessed.WithLabelValues(eventType, status).Inc()
}


package ports

import "time"

type MetricsCollector interface {
	IncrementProductsCreated()
	IncrementProductsDeleted()
	RecordRequestDuration(method, endpoint, status string, duration time.Duration)
	IncrementRequestCount(method, endpoint, status string)
	RecordDatabaseQueryDuration(duration time.Duration)
	RecordRabbitMQPublishDuration(duration time.Duration)
	
	IncrementTransactionRetry()
	
	IncrementTransactionRetrySuccess()
	
	IncrementTransactionRetryFailed()
	
	RecordBatchSize(operation string, size int)
	
	RecordOutboxRetryAttempt(eventType string, attempt int)
	
	RecordOutboxEventProcessed(eventType string, status string)
}


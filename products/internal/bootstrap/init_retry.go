package bootstrap

import (
	"product_service/products/internal/infrastructure/retry"
	"product_service/products/internal/usecase/ports"
)

func initRetrier() ports.Retrier {
	return retry.NewRetryAdapter()
}


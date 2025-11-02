### Products Service
REST API service for managing products with CRUD operations. Features:
- Product creation, retrieval, and deletion
- Pagination support for product listings
- PostgreSQL database with migrations
- Outbox pattern for reliable event publishing
- Prometheus metrics and health checks
- OpenTelemetry distributed tracing

**Port:** 8080  
**Endpoints:**
- `POST /api/v1/products` - Create a product
- `GET /api/v1/products` - List products with pagination
- `DELETE /api/v1/products/:id` - Delete a product
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

### Notifications Service
Event-driven service that consumes product events from RabbitMQ and processes notifications. Features:
- RabbitMQ consumer for product events
- Event processing and logging
- Prometheus metrics and health checks

**Port:** 8081  
**Endpoints:**
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics

## Run services

```bash
docker-compose up -d --build
```

## Stop services

```bash
docker-compose down
```
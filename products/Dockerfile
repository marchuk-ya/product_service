FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY products/go.mod products/go.sum ./
RUN go mod download

COPY products/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o products ./cmd/products

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/products .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./products"]


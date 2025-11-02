FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY notifications/go.mod notifications/go.sum ./
RUN go mod download

COPY notifications/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notifications ./cmd/notifications

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/notifications .

EXPOSE 8081

CMD ["./notifications"]
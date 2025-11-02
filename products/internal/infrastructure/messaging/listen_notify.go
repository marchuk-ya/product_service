package messaging

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgreSQLListener struct {
	db     *sql.DB
	listener *pq.Listener
	logger *zap.Logger
	notify chan struct{}
}

func NewPostgreSQLListener(db *sql.DB, logger *zap.Logger) (*PostgreSQLListener, error) {
	listener := pq.NewListener(
		"",
		10*1000,
		time.Second,
		nil,
	)
	
	
	return &PostgreSQLListener{
		db:       db,
		listener: listener,
		logger:   logger,
		notify:   make(chan struct{}, 1),
	}, nil
}

func (l *PostgreSQLListener) Start(ctx context.Context, channel string) error {
	err := l.listener.Listen(channel)
	if err != nil {
		return fmt.Errorf("failed to listen on channel: %w", err)
	}
	
	l.logger.Info("Started listening for PostgreSQL notifications", zap.String("channel", channel))
	
	go l.handleNotifications(ctx)
	
	return nil
}

func (l *PostgreSQLListener) handleNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case notification := <-l.listener.Notify:
			if notification != nil {
				l.logger.Info("Received PostgreSQL notification",
					zap.String("channel", notification.Channel),
					zap.String("payload", notification.Extra),
				)
				select {
				case l.notify <- struct{}{}:
				default:
				}
			}
		case <-time.After(90 * time.Second):
			l.listener.Ping()
		}
	}
}

func (l *PostgreSQLListener) NotifyChannel() <-chan struct{} {
	return l.notify
}

func (l *PostgreSQLListener) Close() error {
	if l.listener != nil {
		l.listener.Close()
	}
	close(l.notify)
	return nil
}


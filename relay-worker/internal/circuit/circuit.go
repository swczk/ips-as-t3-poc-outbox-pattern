package circuit

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const pingTimeout = 3 * time.Second

type Breaker struct {
	open bool
}

func New() *Breaker { return &Breaker{} }

func (b *Breaker) IsOpen() bool  { return b.open }
func (b *Breaker) IsClosed() bool { return !b.open }

func (b *Breaker) Open() {
	if !b.open {
		b.open = true
		slog.Warn("circuit breaker opened — kafka unavailable")
	}
}

func (b *Breaker) Close() {
	if b.open {
		b.open = false
		slog.Info("circuit breaker closed — kafka recovered")
	}
}

func Ping(brokers string) bool {
	addr := strings.TrimSpace(strings.SplitN(brokers, ",", 2)[0])

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(pingTimeout))
	_, err = conn.ReadPartitions()
	return err == nil
}

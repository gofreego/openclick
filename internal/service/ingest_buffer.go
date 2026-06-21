package service

import (
	"context"
	"sync"
	"time"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/internal/models/dao"
)

// IngestBuffer is a thread-safe in-memory ring buffer for events.
// Events accumulate until the buffer is full or the flush interval elapses.
type IngestBuffer struct {
	mu       sync.Mutex
	buf      []*dao.Event
	cap      int
	analytics AnalyticsRepository
}

// NewIngestBuffer creates and starts the background flush goroutine.
func NewIngestBuffer(ctx context.Context, analytics AnalyticsRepository, capacity int) *IngestBuffer {
	b := &IngestBuffer{
		buf:      make([]*dao.Event, 0, capacity),
		cap:      capacity,
		analytics: analytics,
	}
	go b.flushLoop(ctx)
	return b
}

// Push adds events to the buffer, flushing immediately if capacity is reached.
func (b *IngestBuffer) Push(ctx context.Context, events ...*dao.Event) {
	b.mu.Lock()
	b.buf = append(b.buf, events...)
	shouldFlush := len(b.buf) >= b.cap
	var batch []*dao.Event
	if shouldFlush {
		batch = b.buf
		b.buf = make([]*dao.Event, 0, b.cap)
	}
	b.mu.Unlock()

	if shouldFlush {
		b.flush(ctx, batch)
	}
}

// flushLoop flushes the buffer every second regardless of size.
func (b *IngestBuffer) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Final flush on shutdown
			b.mu.Lock()
			batch := b.buf
			b.buf = make([]*dao.Event, 0, b.cap)
			b.mu.Unlock()
			if len(batch) > 0 {
				b.flush(context.Background(), batch)
			}
			return
		case <-ticker.C:
			b.mu.Lock()
			if len(b.buf) == 0 {
				b.mu.Unlock()
				continue
			}
			batch := b.buf
			b.buf = make([]*dao.Event, 0, b.cap)
			b.mu.Unlock()
			b.flush(ctx, batch)
		}
	}
}

// flush writes a batch of events to ClickHouse.
func (b *IngestBuffer) flush(ctx context.Context, batch []*dao.Event) {
	if len(batch) == 0 {
		return
	}
	if err := b.analytics.InsertEvents(ctx, batch); err != nil {
		logger.Error(ctx, "ingest buffer flush: %v (dropped %d events)", err, len(batch))
	} else {
		logger.Debug(ctx, "flushed %d events to ClickHouse", len(batch))
	}
}

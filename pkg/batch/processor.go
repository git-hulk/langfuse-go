// Package batch provides a generic, type-safe batch processor for efficient API ingestion.
//
// The processor buffers incoming records in a channel-based queue and batches them
// by size or time interval. It uses configurable worker goroutines for parallel
// processing and provides graceful shutdown with timeout handling.
//
// The batch processor is designed to work with any type implementing the Sender[T] interface,
// making it reusable across different API endpoints and data types.
package batch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/git-hulk/langfuse-go/pkg/logger"
)

var (
	ErrProcessorClosed = errors.New("batch processor is closed")
	ErrBufferFull      = errors.New("event buffer is full")
	ErrShutdownTimeout = errors.New("shutdown timeout exceeded")
)

// Sender defines the interface for sending batched records to an external service.
//
// Implementations should handle the actual HTTP requests or other transport mechanisms
// to deliver the batched records. The Send method receives a context for cancellation
// and a slice of records to be sent as a batch.
type Sender[T any] interface {
	Send(ctx context.Context, records []T) error
}

// Config holds the configuration for the batch processor.
type Config struct {
	// MaxBatchSize defines the maximum number of records to send in a single batch.
	// Default is 32.
	MaxBatchSize int
	// FlushInterval defines the interval at which the processor will flush the records
	// even if the batch size is not reached.
	// Default is 3 seconds.
	FlushInterval time.Duration
	// BufferSize defines the size of the internal buffer for incoming records.
	// If the buffer is full, Submit will return an error.
	// Default is MaxBatchSize * 10.
	BufferSize int
	// NumWorkers defines the number of worker goroutines that will process the batches.
	// Default is 1.
	NumWorkers int
	// ShutdownTimeout defines the maximum time to wait for the processor to shut down gracefully.
	// If the processor does not shut down within this time, an error will be returned.
	// Default is 30 seconds.
	ShutdownTimeout time.Duration
}

func (c *Config) normalize() {
	if c.FlushInterval <= 0 {
		c.FlushInterval = 3 * time.Second
	}
	if c.MaxBatchSize <= 0 {
		c.MaxBatchSize = 32
	}
	if c.BufferSize <= 0 {
		c.BufferSize = c.MaxBatchSize * 10
	}
	if c.ShutdownTimeout <= 0 {
		c.ShutdownTimeout = 30 * time.Second
	}
	if c.NumWorkers <= 0 {
		c.NumWorkers = 1
	}
}

func defaultConfig() *Config {
	return &Config{
		MaxBatchSize:    100,
		FlushInterval:   3 * time.Second,
		BufferSize:      1000,
		ShutdownTimeout: 30 * time.Second,
		NumWorkers:      1,
	}
}

// Processor is a generic, type-safe batch processor that efficiently collects and sends records.
//
// The processor uses a channel-based architecture with configurable batching by size and time.
// It supports multiple worker goroutines for parallel processing and provides graceful shutdown
// with timeout handling. Records are buffered in memory and automatically flushed when batch
// size limits are reached or flush intervals expire.
//
// The processor is thread-safe and can be used concurrently from multiple goroutines.
type Processor[T any] struct {
	config  *Config
	sender  Sender[T]
	buffer  chan T
	pending chan []T

	quitCh chan struct{}

	wg     sync.WaitGroup
	closed atomic.Bool
}

type applyOption func(*Config)

// NewProcessor creates a new Processor instance with the provided Sender and optional configuration.
//
// The processor is immediately started with the configured number of worker goroutines.
// Use the provided With* option functions to customize batch size, flush interval,
// buffer size, number of workers, and shutdown timeout.
//
// Example:
//
//	processor := NewProcessor(sender,
//		WithMaxBatchSize(50),
//		WithFlushInterval(5*time.Second),
//		WithNumWorkers(2),
//	)
func NewProcessor[T any](sender Sender[T], options ...applyOption) *Processor[T] {
	config := defaultConfig()
	for _, opt := range options {
		opt(config)
	}
	config.normalize()

	p := &Processor[T]{
		config:  config,
		sender:  sender,
		buffer:  make(chan T, config.BufferSize),
		pending: make(chan []T, config.NumWorkers*2),
		quitCh:  make(chan struct{}),
	}

	ctx := context.Background()
	p.wg.Add(1 + config.NumWorkers)
	go p.collectRecords(ctx)

	for i := 0; i < config.NumWorkers; i++ {
		go p.sendBatchLoop(ctx)
	}

	return p
}

// WithMaxBatchSize sets the maximum number of records to send in a single batch.
// Default is 100 records per batch.
func WithMaxBatchSize(maxBatchSize int) applyOption {
	return func(c *Config) {
		c.MaxBatchSize = maxBatchSize
	}
}

// WithFlushInterval sets the time interval for automatic batch flushing.
// Batches will be sent after this interval even if not full. Default is 3 seconds.
func WithFlushInterval(flushInterval time.Duration) applyOption {
	return func(c *Config) {
		c.FlushInterval = flushInterval
	}
}

// WithBufferSize sets the size of the internal record buffer.
// If the buffer is full, Submit will return an error. Default is 1000 records.
func WithBufferSize(bufferSize int) applyOption {
	return func(c *Config) {
		c.BufferSize = bufferSize
	}
}

// WithNumWorkers sets the number of worker goroutines for processing batches.
// More workers enable higher concurrency but use more resources. Default is 1.
func WithNumWorkers(numWorkers int) applyOption {
	return func(c *Config) {
		c.NumWorkers = numWorkers
	}
}

// WithShutdownTimeout sets the maximum time to wait for graceful shutdown.
// If the processor doesn't shut down within this time, an error is returned. Default is 30 seconds.
func WithShutdownTimeout(shutdownTimeout time.Duration) applyOption {
	return func(c *Config) {
		c.ShutdownTimeout = shutdownTimeout
	}
}

// Submit adds a record to the processor's buffer. If the buffer is full, it returns an error.
func (p *Processor[T]) Submit(record T) error {
	if p.closed.Load() {
		return ErrProcessorClosed
	}

	select {
	case p.buffer <- record:
		return nil
	default:
		return ErrBufferFull
	}
}

// Close gracefully shuts down the processor, ensuring all pending records are sent.
// It waits for the shutdown to complete or times out based on the configured ShutdownTimeout.
func (p *Processor[T]) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil
	}

	close(p.quitCh)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(p.config.ShutdownTimeout):
		return ErrShutdownTimeout
	}
}

func (p *Processor[T]) collectRecords(ctx context.Context) {
	defer p.wg.Done()

	tick := time.NewTicker(p.config.FlushInterval)
	defer tick.Stop()

	batchBuffer := make([]T, 0, p.config.MaxBatchSize)

	for {
		select {
		case record := <-p.buffer:
			batchBuffer = append(batchBuffer, record)
			if len(batchBuffer) >= p.config.MaxBatchSize {
				p.sendBatch(ctx, batchBuffer)
				batchBuffer = batchBuffer[:0]
			}
		case <-tick.C:
			if len(batchBuffer) > 0 {
				p.sendBatch(ctx, batchBuffer)
				batchBuffer = batchBuffer[:0]
			}
		case <-p.quitCh:
			for len(p.buffer) > 0 {
				record := <-p.buffer
				batchBuffer = append(batchBuffer, record)
				if len(batchBuffer) >= p.config.MaxBatchSize {
					p.sendBatch(ctx, batchBuffer)
					batchBuffer = batchBuffer[:0]
				}
			}
			if len(batchBuffer) > 0 {
				p.sendBatch(ctx, batchBuffer)
			}
			close(p.pending)
			return
		}
	}
}

func (p *Processor[T]) sendBatchLoop(ctx context.Context) {
	defer p.wg.Done()

	for {
		select {
		case batch, ok := <-p.pending:
			if !ok {
				return
			}
			p.sendBatch(ctx, batch)
		case <-p.quitCh:
			for batch := range p.pending {
				p.sendBatch(ctx, batch)
			}
			return
		}
	}
}

func (p *Processor[T]) sendBatch(ctx context.Context, records []T) {
	if len(records) == 0 {
		return
	}
	if err := p.sender.Send(ctx, records); err != nil {
		logger.Get().Error("Failed to send batch", zap.Error(err))
	}
}

package batch

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockSender struct {
	mu        sync.Mutex
	batches   [][]any
	sendCount int
	failAfter int
	sendDelay time.Duration
}

func (m *mockSender) Send(_ context.Context, events []any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendDelay > 0 {
		time.Sleep(m.sendDelay)
	}

	m.sendCount++
	if m.failAfter > 0 && m.sendCount > m.failAfter {
		return errors.New("mock send failure")
	}

	m.batches = append(m.batches, append([]any(nil), events...))
	return nil
}

func (m *mockSender) getBatches() [][]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([][]any(nil), m.batches...)
}

func (m *mockSender) getSendCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sendCount
}

type countingSender struct {
	count *int64
}

func (s *countingSender) Send(_ context.Context, _ []any) error {
	atomic.AddInt64(s.count, 1)
	return nil
}

type concurrencyTrackingSender struct {
	active        *int64
	maxConcurrent *int64
	delay         time.Duration
}

func (s *concurrencyTrackingSender) Send(_ context.Context, _ []any) error {
	atomic.AddInt64(s.active, 1)
	cur := atomic.LoadInt64(s.active)
	maxVal := atomic.LoadInt64(s.maxConcurrent)
	if cur > maxVal {
		atomic.StoreInt64(s.maxConcurrent, cur)
	}
	time.Sleep(s.delay)
	atomic.AddInt64(s.active, -1)
	return nil
}

func TestProcessor_Submit(t *testing.T) {
	sender := &mockSender{}
	processor := NewProcessor[any](sender,
		WithMaxBatchSize(3),
		WithBufferSize(10),
		WithFlushInterval(time.Millisecond),
	)
	defer processor.Close()

	for i := 0; i < 6; i++ {
		require.NoError(t, processor.Submit(i))
	}

	time.Sleep(5 * time.Millisecond)

	batches := sender.getBatches()
	require.GreaterOrEqual(t, len(batches), 2)

	totalEvents := 0
	for _, batch := range batches {
		totalEvents += len(batch)
	}
	require.Equal(t, 6, totalEvents)
}

func TestProcessor_MaxBatchSize(t *testing.T) {
	sender := &mockSender{}
	processor := NewProcessor[any](sender, WithMaxBatchSize(2))
	defer processor.Close()

	for i := 0; i < 5; i++ {
		processor.Submit(i)
	}

	time.Sleep(100 * time.Millisecond)

	batches := sender.getBatches()
	require.GreaterOrEqual(t, len(batches), 2)
}

func TestProcessor_MultipleWorkers(t *testing.T) {
	var sendCount int64
	sender := &countingSender{count: &sendCount}
	processor := NewProcessor[any](sender,
		WithNumWorkers(4))
	defer processor.Close()

	numEvents := 100
	for i := 0; i < numEvents; i++ {
		require.NoError(t, processor.Submit(i))
	}

	time.Sleep(300 * time.Millisecond)

	totalSent := atomic.LoadInt64(&sendCount)
	require.Greater(t, totalSent, int64(0))
}

func TestProcessor_MultiWorker(t *testing.T) {
	sender := &mockSender{}
	processor := NewProcessor[any](sender,
		WithMaxBatchSize(3),
		WithBufferSize(120),
		WithNumWorkers(4),
	)
	defer processor.Close()

	for i := 0; i < 60; i++ {
		require.NoError(t, processor.Submit(i))
	}

	time.Sleep(100 * time.Millisecond)

	batches := sender.getBatches()
	require.Greater(t, len(batches), 0)

	totalEvents := 0
	for _, batch := range batches {
		totalEvents += len(batch)
	}
	require.Equal(t, 60, totalEvents)
}

func TestProcessor_SingleWorker(t *testing.T) {
	sender := &mockSender{}
	processor := NewProcessor[any](sender,
		WithMaxBatchSize(3),
		WithBufferSize(10),
		WithNumWorkers(1),
	)
	defer processor.Close()

	for i := 0; i < 6; i++ {
		require.NoError(t, processor.Submit(i))
	}

	time.Sleep(100 * time.Millisecond)

	batches := sender.getBatches()
	require.Greater(t, len(batches), 0)

	totalEvents := 0
	for _, batch := range batches {
		totalEvents += len(batch)
	}
	require.Equal(t, 6, totalEvents)
}

func TestProcessor_Close(t *testing.T) {
	sender := &mockSender{}
	processor := NewProcessor[any](sender,
		WithMaxBatchSize(10),
		WithBufferSize(10),
		WithShutdownTimeout(5*time.Second),
	)

	require.NoError(t, processor.Submit("event1"))
	require.NoError(t, processor.Submit("event2"))

	require.NoError(t, processor.Close())

	batches := sender.getBatches()
	require.Equal(t, 1, len(batches))

	err := processor.Submit("event3")
	require.Equal(t, ErrProcessorClosed, err)
}

func TestProcessor_BufferFull(t *testing.T) {
	sender := &mockSender{sendDelay: 100 * time.Millisecond}
	processor := NewProcessor[any](sender,
		WithMaxBatchSize(10),
		WithBufferSize(2),
	)
	defer func() { require.NoError(t, processor.Close()) }()

	require.NoError(t, processor.Submit("event1"))
	require.NoError(t, processor.Submit("event2"))

	err := processor.Submit("event3")
	require.Equal(t, ErrBufferFull, err)
}

func TestProcessor_DefaultConfig(t *testing.T) {
	sender := &mockSender{}
	processor := NewProcessor(sender)
	defer func() { require.NoError(t, processor.Close()) }()

	require.Equal(t, 100, processor.config.MaxBatchSize)
}

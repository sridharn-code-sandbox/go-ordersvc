package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/nsridhar76/go-ordersvc/internal/domain"
	"github.com/nsridhar76/go-ordersvc/internal/messaging"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWriter captures messages written to Kafka for test assertions.
type mockWriter struct {
	mu       sync.Mutex
	messages []kafkago.Message
	err      error
	closed   bool
}

func (m *mockWriter) WriteMessages(_ context.Context, msgs ...kafkago.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *mockWriter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockWriter) lastMessage() kafkago.Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages[len(m.messages)-1]
}

func newTestPublisher(w *mockWriter) *Publisher {
	return &Publisher{writer: w, topic: "order-events"}
}

func newTestOrder() *domain.Order {
	return &domain.Order{
		ID:         uuid.New(),
		CustomerID: "cust-123",
		Items: []domain.OrderItem{
			{ID: uuid.New(), ProductID: "p-1", Name: "Widget", Quantity: 2, Price: 10.50, Subtotal: 21.00},
		},
		Status:  domain.OrderStatusPending,
		Total:   21.00,
		Version: 1,
	}
}

func TestPublisher_PublishOrderCreated_WritesCorrectMessage(t *testing.T) {
	w := &mockWriter{}
	pub := newTestPublisher(w)
	order := newTestOrder()

	err := pub.PublishOrderCreated(context.Background(), order)

	require.NoError(t, err)
	assert.Len(t, w.messages, 1)

	msg := w.lastMessage()
	assert.Equal(t, order.ID.String(), string(msg.Key))

	var evt messaging.OrderEvent
	require.NoError(t, json.Unmarshal(msg.Value, &evt))
	assert.Equal(t, messaging.EventOrderCreated, evt.EventType)
	assert.Equal(t, order.ID.String(), evt.OrderID)
	assert.Equal(t, order.CustomerID, evt.CustomerID)
	assert.Equal(t, "pending", evt.Status)
	assert.Equal(t, 21.00, evt.Total)
	assert.Equal(t, 1, evt.Version)
	assert.Empty(t, evt.OldStatus)
	assert.Empty(t, evt.NewStatus)
	assert.False(t, evt.OccurredAt.IsZero())
}

func TestPublisher_PublishOrderUpdated_WritesCorrectMessage(t *testing.T) {
	w := &mockWriter{}
	pub := newTestPublisher(w)
	order := newTestOrder()
	order.Status = domain.OrderStatusConfirmed
	order.Version = 3

	err := pub.PublishOrderUpdated(context.Background(), order)

	require.NoError(t, err)
	msg := w.lastMessage()
	assert.Equal(t, order.ID.String(), string(msg.Key))

	var evt messaging.OrderEvent
	require.NoError(t, json.Unmarshal(msg.Value, &evt))
	assert.Equal(t, messaging.EventOrderUpdated, evt.EventType)
	assert.Equal(t, "confirmed", evt.Status)
	assert.Equal(t, 3, evt.Version)
}

func TestPublisher_PublishOrderStatusChanged_IncludesOldAndNewStatus(t *testing.T) {
	w := &mockWriter{}
	pub := newTestPublisher(w)
	order := newTestOrder()
	order.Status = domain.OrderStatusConfirmed

	err := pub.PublishOrderStatusChanged(
		context.Background(),
		order,
		domain.OrderStatusPending,
		domain.OrderStatusConfirmed,
	)

	require.NoError(t, err)
	msg := w.lastMessage()

	var evt messaging.OrderEvent
	require.NoError(t, json.Unmarshal(msg.Value, &evt))
	assert.Equal(t, messaging.EventOrderStatusChanged, evt.EventType)
	assert.Equal(t, "pending", evt.OldStatus)
	assert.Equal(t, "confirmed", evt.NewStatus)
	assert.Equal(t, "confirmed", evt.Status)
}

func TestPublisher_PublishOrderCreated_WriterError_ReturnsError(t *testing.T) {
	w := &mockWriter{err: errors.New("broker unavailable")}
	pub := newTestPublisher(w)

	err := pub.PublishOrderCreated(context.Background(), newTestOrder())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "broker unavailable")
}

func TestPublisher_Close_ClosesWriter(t *testing.T) {
	w := &mockWriter{}
	pub := newTestPublisher(w)

	err := pub.Close()

	assert.NoError(t, err)
	assert.True(t, w.closed)
}

func TestPublisher_MessageKey_IsOrderID(t *testing.T) {
	tests := []struct {
		name    string
		publish func(pub *Publisher, order *domain.Order) error
	}{
		{
			name: "created",
			publish: func(pub *Publisher, order *domain.Order) error {
				return pub.PublishOrderCreated(context.Background(), order)
			},
		},
		{
			name: "updated",
			publish: func(pub *Publisher, order *domain.Order) error {
				return pub.PublishOrderUpdated(context.Background(), order)
			},
		},
		{
			name: "status_changed",
			publish: func(pub *Publisher, order *domain.Order) error {
				return pub.PublishOrderStatusChanged(
					context.Background(), order,
					domain.OrderStatusPending, domain.OrderStatusConfirmed,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &mockWriter{}
			pub := newTestPublisher(w)
			order := newTestOrder()

			err := tt.publish(pub, order)
			require.NoError(t, err)

			msg := w.lastMessage()
			assert.Equal(t, order.ID.String(), string(msg.Key),
				"message key must be order ID for partition ordering")
		})
	}
}

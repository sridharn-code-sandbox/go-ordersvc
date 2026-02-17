//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nsridhar76/go-ordersvc/internal/messaging"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func kafkaBroker() string {
	if b := os.Getenv("KAFKA_BROKERS"); b != "" {
		return b
	}
	return "localhost:29092"
}

const kafkaTopic = "order-events"

// findEvent scans the topic from the beginning looking for an event matching the predicate.
// This avoids consumer-group rebalance timing issues in tests.
func findEvent(t *testing.T, timeout time.Duration, match func(messaging.OrderEvent) bool) (messaging.OrderEvent, kafka.Message) {
	t.Helper()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker()},
		Topic:     kafkaTopic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  1e6,
	})
	reader.SetOffset(kafka.FirstOffset)
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				t.Fatalf("timed out after %v waiting for matching Kafka event", timeout)
			}
			t.Fatalf("failed to read Kafka message: %v", err)
		}

		var evt messaging.OrderEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			continue
		}
		if match(evt) {
			return evt, msg
		}
	}
}

// findEvents collects all events matching the predicate within the timeout.
func findEvents(t *testing.T, timeout time.Duration, match func(messaging.OrderEvent) bool) []messaging.OrderEvent {
	t.Helper()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaBroker()},
		Topic:     kafkaTopic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  1e6,
	})
	reader.SetOffset(kafka.FirstOffset)
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var events []messaging.OrderEvent
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			// timeout means we've consumed everything available
			return events
		}

		var evt messaging.OrderEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			continue
		}
		if match(evt) {
			events = append(events, evt)
		}
	}
}

// =============================================================================
// Integration Tests: Kafka Event Publishing (ADR-0006)
// =============================================================================

func TestKafka_CreateOrder_PublishesOrderCreatedEvent(t *testing.T) {
	customerID := uuid.New().String()

	// Create order via HTTP API
	req := CreateOrderRequest{
		CustomerID: customerID,
		Items: []OrderItem{
			{ProductID: "kafka-test-1", Name: "Kafka Test Product", Quantity: 3, Price: 15.00},
		},
	}
	resp, body := post(t, "/api/v1/orders", req)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created OrderResponse
	require.NoError(t, json.Unmarshal(body, &created))

	// Find the created event in Kafka
	evt, _ := findEvent(t, 15*time.Second, func(e messaging.OrderEvent) bool {
		return e.OrderID == created.ID && e.EventType == messaging.EventOrderCreated
	})

	assert.Equal(t, messaging.EventOrderCreated, evt.EventType)
	assert.Equal(t, created.ID, evt.OrderID)
	assert.Equal(t, customerID, evt.CustomerID)
	assert.Equal(t, "pending", evt.Status)
	assert.Equal(t, 45.00, evt.Total)
	assert.Empty(t, evt.OldStatus, "order.created should have no old_status")
	assert.Empty(t, evt.NewStatus, "order.created should have no new_status")
	assert.False(t, evt.OccurredAt.IsZero())
}

func TestKafka_UpdateOrderStatus_PublishesStatusChangedEvent(t *testing.T) {
	// Create an order first
	createReq := CreateOrderRequest{
		CustomerID: uuid.New().String(),
		Items: []OrderItem{
			{ProductID: "kafka-status-1", Name: "Status Test", Quantity: 1, Price: 25.00},
		},
	}
	createResp, createBody := post(t, "/api/v1/orders", createReq)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created OrderResponse
	require.NoError(t, json.Unmarshal(createBody, &created))

	// Update status: pending -> confirmed
	statusResp, _ := patch(t, "/api/v1/orders/"+created.ID+"/status", UpdateStatusRequest{Status: "confirmed"})
	require.Equal(t, http.StatusOK, statusResp.StatusCode)

	// Find the status changed event in Kafka
	evt, _ := findEvent(t, 15*time.Second, func(e messaging.OrderEvent) bool {
		return e.OrderID == created.ID && e.EventType == messaging.EventOrderStatusChanged
	})

	assert.Equal(t, messaging.EventOrderStatusChanged, evt.EventType)
	assert.Equal(t, created.ID, evt.OrderID)
	assert.Equal(t, "confirmed", evt.Status)
	assert.Equal(t, "pending", evt.OldStatus)
	assert.Equal(t, "confirmed", evt.NewStatus)
}

func TestKafka_OrderLifecycle_ProducesOrderedEvents(t *testing.T) {
	customerID := uuid.New().String()

	// 1. Create order
	createReq := CreateOrderRequest{
		CustomerID: customerID,
		Items: []OrderItem{
			{ProductID: "lifecycle-1", Name: "Lifecycle Product", Quantity: 1, Price: 100.00},
		},
	}
	createResp, createBody := post(t, "/api/v1/orders", createReq)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created OrderResponse
	require.NoError(t, json.Unmarshal(createBody, &created))

	// 2. Transition: pending -> confirmed -> processing
	statusResp, _ := patch(t, "/api/v1/orders/"+created.ID+"/status", UpdateStatusRequest{Status: "confirmed"})
	require.Equal(t, http.StatusOK, statusResp.StatusCode)

	statusResp, _ = patch(t, "/api/v1/orders/"+created.ID+"/status", UpdateStatusRequest{Status: "processing"})
	require.Equal(t, http.StatusOK, statusResp.StatusCode)

	// Give Kafka a moment to flush
	time.Sleep(500 * time.Millisecond)

	// Find all events for this order
	events := findEvents(t, 15*time.Second, func(e messaging.OrderEvent) bool {
		return e.OrderID == created.ID
	})

	require.Len(t, events, 3, "should have 3 events: created + 2 status changes")

	assert.Equal(t, messaging.EventOrderCreated, events[0].EventType, "first event should be order.created")

	assert.Equal(t, messaging.EventOrderStatusChanged, events[1].EventType, "second event should be status_changed")
	assert.Equal(t, "pending", events[1].OldStatus)
	assert.Equal(t, "confirmed", events[1].NewStatus)

	assert.Equal(t, messaging.EventOrderStatusChanged, events[2].EventType, "third event should be status_changed")
	assert.Equal(t, "confirmed", events[2].OldStatus)
	assert.Equal(t, "processing", events[2].NewStatus)
}

func TestKafka_MessageKey_IsOrderID(t *testing.T) {
	// Create order
	createReq := CreateOrderRequest{
		CustomerID: uuid.New().String(),
		Items: []OrderItem{
			{ProductID: "key-test-1", Name: "Key Test", Quantity: 1, Price: 50.00},
		},
	}
	createResp, createBody := post(t, "/api/v1/orders", createReq)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created OrderResponse
	require.NoError(t, json.Unmarshal(createBody, &created))

	// Find the message and check its key
	_, msg := findEvent(t, 15*time.Second, func(e messaging.OrderEvent) bool {
		return e.OrderID == created.ID && e.EventType == messaging.EventOrderCreated
	})

	assert.Equal(t, created.ID, string(msg.Key),
		"Kafka message key must be order ID for per-order partition ordering (ADR-0006)")
}

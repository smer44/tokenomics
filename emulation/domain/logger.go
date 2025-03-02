package domain

import (
	"log/slog"
	"os"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// Log domain events with structured data
func logEvent(event string, attrs ...slog.Attr) {
	// Convert []slog.Attr to []any
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	logger.Info(event, args...)
}

// Helper functions to create common attributes
func withOrderId(id OrderId) slog.Attr {
	return slog.String("orderId", string(id))
}

func withProducerId(id ProducerId) slog.Attr {
	return slog.String("producerId", string(id))
}

func withConsumerId(id ConsumerId) slog.Attr {
	return slog.String("consumerId", string(id))
}

func withTokens(tokens Tokens) slog.Attr {
	return slog.Int("tokens", int(tokens))
}

func withCapacity(capacity Capacity) slog.Attr {
	return slog.Int("capacity", int(capacity))
}

func withCapacityType(ct CapacityType) slog.Attr {
	return slog.String("capacityType", string(ct))
}

func withProduct(product Product) slog.Attr {
	return slog.Int("product", int(product))
}

func withCutOffPrice(price CapacityUnitPrice) slog.Attr {
	return slog.Float64("cutOffPrice", float64(price))
}

func withCycleCounter(counter uint) slog.Attr {
	return slog.Uint64("cycleCounter", uint64(counter))
}

func withState(state SystemState) slog.Attr {
	return slog.String("state", systemStateToString(state))
}

func systemStateToString(state SystemState) string {
	switch state {
	case SystemStateOrdersPlacement:
		return "OrdersPlacement"
	case SystemStateOrdering:
		return "Ordering"
	default:
		return "Unknown"
	}
}

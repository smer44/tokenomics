package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrder(t *testing.T) {
	ps := ProcessSheet{
		Product: 1,
		Require: map[CapacityType]Capacity{
			"1": 10,
			"2": 20},
	}
	consRequest := ConsumerRequest{"1", 1, 100}
	investementRequest := InvestmentRequest{"1", InvestmentTypeUpgrade, 1, 30}

	t.Run(`Given a customer order
		When denied operations are called
		Then should panics`, func(t *testing.T) {

		order := NewConsumerOrder("1", ps, consRequest)
		require.Panics(t, func() { order.CutOffPrice() })
		require.Panics(t, func() { order.Fund(100) })
		require.Panics(t, func() { order.Processing("3", 100) })
		require.Panics(t, func() { order.Rejected("3") })
		require.Panics(t, func() { order.Completed("3", 100) })
		// wrong token numbers
		require.Panics(t, func() { order.Processing("1", 101) })
		require.Panics(t, func() { order.Completed("1", 101) })
	})

	t.Run(`Given a investement order
		When denied operations are called
		Then should panics`, func(t *testing.T) {

		order := NewInvestmentOrder("1", ps, investementRequest)
		require.Panics(t, func() { order.Processing("3", 100) })
		require.Panics(t, func() { order.Rejected("3") })
		require.Panics(t, func() { order.Completed("3", 100) })
		require.Panics(t, func() { order.CompleteCycle() })
		require.Panics(t, func() { order.Info() })
		require.Panics(t, func() { order.AgentId() })
	})

	t.Run(`Given a investement order
		When Rejected is called for every bid
		And CompleteCycle is called
		Then should return ConsumerRequestRejected`, func(t *testing.T) {

		order := NewInvestmentOrder("1", ps, investementRequest)
		require.True(t, order.RequiresFunding())
		order.Fund(200)
		require.False(t, order.RequiresFunding())
		order.Rejected("1")
		order.Rejected("2")
		score, event := order.CompleteCycle()
		require.Equal(t, Score(3), score)
		require.Equal(t, InvestmentRequestRejected{&investementRequest}, event)
	})

	t.Run(`Given a investement order
		When Completed is called for every bid
		And CompleteCycle is called
		Then should return ConsumerRequestRejected`, func(t *testing.T) {

		order := NewInvestmentOrder("1", ps, investementRequest)
		order.Fund(200)
		order.Info()

		order.Completed("1", 100)
		order.Completed("2", 100)
		score, event := order.CompleteCycle()
		require.Equal(t, Score(0), score)
		require.Equal(t, InvestmentRequestCompleted{&investementRequest}, event)
	})

	t.Run(`Given a customer order
		When Rejected is called for every bid
		And CompleteCycle is called
		Then should return ConsumerRequestRejected`, func(t *testing.T) {
		order := NewConsumerOrder("1", ps, consRequest)
		order.Rejected("1")
		order.Rejected("2")
		score, event := order.CompleteCycle()
		require.Equal(t, Score(3), score)
		require.Equal(t, ConsumerRequestRejected{100, &consRequest}, event)
	})

	t.Run(`Given a customer order
		When Completed is called for every bid
		And CompleteCycle is called
		Then should return ConsumerRequestCompleted`, func(t *testing.T) {

		order := NewConsumerOrder("1", ps, consRequest)
		order.Completed("1", 50)
		order.Completed("2", 50)
		score, event := order.CompleteCycle()
		require.Equal(t, Score(0), score)
		require.Equal(t, ConsumerRequestCompleted{&consRequest}, event)
	})

	t.Run(`Given a customer order
		And one bid is processing with 50 tokens
		And another is rejected 3 cycles
		When CompleteCycle is called
		Then should return ConsumerRequestRejected
		And tokens should
		`, func(t *testing.T) {
		order := NewConsumerOrder("1", ps, consRequest)
		order.Processing("1", 50)
		var score Score
		var event OrderEvent
		for i := 0; i < 3; i++ {
			order.Rejected("2")
			score, event = order.CompleteCycle()
		}
		require.Equal(t, Score(5), score)
		require.Equal(t, ConsumerRequestRejected{50, &consRequest}, event)
	})

	t.Run(`Given a customer order
		And one bid is completed with 50 tokens
		And another is rejected 3 cycles
		When CompleteCycle is called
		Then should return ConsumerRequestRejected
		And tokens should
	`, func(t *testing.T) {
		order := NewConsumerOrder("1", ps, consRequest)
		order.Completed("1", 50)
		var score Score
		var event OrderEvent
		for i := 0; i < 3; i++ {
			order.Rejected("2")
			score, event = order.CompleteCycle()
		}
		require.Equal(t, Score(5), score)
		require.Equal(t, ConsumerRequestRejected{50, &consRequest}, event)
	})

	t.Run(`Given a customer order
		And no bids processed
		When CompleteCycle is called
		Then should return StillProcessing`, func(t *testing.T) {

		order := NewConsumerOrder("1", ps, consRequest)
		score, event := order.CompleteCycle()
		require.Equal(t, Score(1), score)
		require.Equal(t, OrderStillProcessing{}, event)
	})

	t.Run(`Given a customer order
		And one bid completed
		When CompleteCycle is called
		Then should return StillProcessing`, func(t *testing.T) {

		order := NewConsumerOrder("1", ps, consRequest)
		order.Completed("1", 10)
		score, event := order.CompleteCycle()
		require.Equal(t, Score(1), score)
		require.Equal(t, OrderStillProcessing{}, event)
	})

	t.Run(`Given a customer order
		And one bid rejected 
		When CompleteCycle is called
		Then should return StillProcessing`, func(t *testing.T) {

		order := NewConsumerOrder("1", ps, consRequest)
		order.Rejected("1")
		score, event := order.CompleteCycle()
		require.Equal(t, Score(1), score)
		require.Equal(t, OrderStillProcessing{}, event)
	})

	t.Run(`Given a customer order
		And one bid completed
		When CompleteCycle is called
		Then should return StillProcessing`, func(t *testing.T) {

		order := NewConsumerOrder("1", ps, consRequest)
		order.Completed("1", 10)
		score, event := order.CompleteCycle()
		require.Equal(t, Score(1), score)
		require.Equal(t, OrderStillProcessing{}, event)
	})

	t.Run(`Given a customer order
		And bid eventually completed
		When Info is called
		Then should return unassigned bids`, func(t *testing.T) {
		order := NewConsumerOrder("1", ps, consRequest)
		require.Equal(t, OrderInfo{"1", 100, map[CapacityType]Capacity{"1": 10, "2": 20}}, order.Info())
		order.Rejected("1")
		order.Rejected("2")
		require.Equal(t, OrderInfo{"1", 100, map[CapacityType]Capacity{"1": 10, "2": 20}}, order.Info())
		order.Processing("1", 10)
		require.Equal(t, OrderInfo{"1", 90, map[CapacityType]Capacity{"2": 20}}, order.Info())
		order.Completed("1", 10)
		require.Equal(t, OrderInfo{"1", 90, map[CapacityType]Capacity{"2": 20}}, order.Info())
		order.Completed("2", 90)
		require.Equal(t, OrderInfo{"1", 0, map[CapacityType]Capacity{}}, order.Info())
	})
}

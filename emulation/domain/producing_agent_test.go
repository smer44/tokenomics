package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func agent(inProgress *Booking, cutOffPrice CapacityUnitPrice) *ProducingAgent {
	pa := newProducingAgent(ProducingAgentConfig{
		"p1", 1, 100, 0, Restoration{}, Upgrade{},
	})
	pa.state.inProgress = inProgress
	pa.state.cutOffPrice = cutOffPrice
	return pa
}

func TestProducingAgent(t *testing.T) {
	type test struct {
		inProgress  *Booking
		cutOffPrice CapacityUnitPrice
		bids        []Bid
		result      ProductionResult
		state       producerState
	}
	tests := []test{
		{nil, 0, nil, ProductionResult{[]OrderId{}, []OrderId{}}, producerState{100, nil, nil, 0, 0, 0}},
		{nil, 0, []Bid{{100, 20, "1"}}, ProductionResult{[]OrderId{"1"}, []OrderId{}}, producerState{100, nil, nil, 100, 20, 0.2}},
		{nil, 0, []Bid{{40, 20, "1"}, {40, 4, "2"}, {20, 20, "3"}, {40, 1, "4"}}, ProductionResult{[]OrderId{"3", "1", "2"}, []OrderId{"4"}}, producerState{100, nil, nil, 140, 44, 0.1}},
		{nil, 0, []Bid{{100, 100, "1"}, {10, 1, "2"}}, ProductionResult{[]OrderId{"1"}, []OrderId{"2"}}, producerState{100, nil, nil, 110, 100, 1}},
		{nil, 0, []Bid{{200, 20, "1"}}, ProductionResult{[]OrderId{}, []OrderId{}}, producerState{100, nil, &Booking{"1", 100}, 200, 20, 0.1}},
		{&Booking{"1", 100}, 3, nil, ProductionResult{[]OrderId{"1"}, []OrderId{}}, producerState{100, nil, nil, 0, 0, 3}},
		{&Booking{"1", 100}, 3, []Bid{{50, 20, "2"}}, ProductionResult{[]OrderId{"1"}, []OrderId{"2"}}, producerState{100, nil, nil, 50, 0, 3}},
		{&Booking{"1", 50}, 3, []Bid{{100, 20, "2"}}, ProductionResult{[]OrderId{"1"}, []OrderId{}}, producerState{100, nil, &Booking{"2", 50}, 100, 20, 0.2}},
	}
	for idx, tc := range tests {
		// if idx != 3 {
		// 	continue
		// }
		t.Run(fmt.Sprintf("[TEST CASE]: %d\n", idx), func(t *testing.T) {
			p := agent(tc.inProgress, tc.cutOffPrice)
			p.PlaceBids(tc.bids)
			result := p.Produce()
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.state, p.state)
		})
	}
}

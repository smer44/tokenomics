package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type option func(*ProducingAgent)

func withCapacity(c Capacity) option {
	return func(pa *ProducingAgent) {
		pa.producerState.capacity = c
	}
}

func withInProgres(ip *Booking) option {
	return func(pa *ProducingAgent) {
		pa.producerState.inProgress = ip
	}
}

func withCutOffPrice(c CapacityUnitPrice) option {
	return func(pa *ProducingAgent) {
		pa.producerState.cutOffPrice = c
	}
}

func withFunds(t Tokens) option {
	return func(pa *ProducingAgent) {
		pa.producerState.funds = t
	}
}

func agent(options ...option) *ProducingAgent {
	pa := newProducingAgent(ProducingAgentConfig{
		"p1", 1, 100, 0, Restoration{ProcessSheet{1, nil}, 7}, Upgrade{ProcessSheet{1, nil}, 10},
	})
	for _, opt := range options {
		opt(pa)
	}
	return pa
}

func TestProducingAgentInvest(t *testing.T) {
	t.Run(`Given producing agent
		When restore is requested
		And restore is done
		Then increase capacity
		But not greater than maxCapacity`, func(t *testing.T) {
		p := agent(withFunds(100), withCapacity(95))
		u, r := p.Invest(ProducingAgentCommand{100, false, true})
		require.Nil(t, u)
		require.Equal(t, RestoreRequest{100, 100, p.restoration}, *r)
		p.CompleteRestore()
		require.Equal(t, Capacity(100), p.producerState.maxCapacity)
		require.Equal(t, Capacity(100), p.producerState.capacity)
		require.Panics(t, func() {
			p.CompleteRestore()
		})
	})

	t.Run(`Given producing agent
		When upgrade is requests
		And upgrade is done
		Then increase max capacity`, func(t *testing.T) {
		p := agent(withFunds(100), withCapacity(50))
		u, r := p.Invest(ProducingAgentCommand{100, true, false})
		require.Nil(t, r)
		require.Equal(t, UpgradeRequest{100, 100, p.upgrade}, *u)
		p.CompleteUpgrade()
		require.Equal(t, Capacity(110), p.producerState.maxCapacity)
		require.Equal(t, Capacity(60), p.producerState.capacity)
		require.Panics(t, func() {
			p.CompleteUpgrade()
		})
	})

	t.Run(`Given a producing agent
		When restore is requested
		And upgrade is requested
		And restore is done
		And upgrade is done
		Then increase maxCapacity
		And restores capacity
		But not greater than maxCapacity`, func(t *testing.T) {
		p := agent(withFunds(100), withCapacity(50))
		u, r := p.Invest(ProducingAgentCommand{100, true, true})
		require.Equal(t, RestoreRequest{50, 50, p.restoration}, *r)
		require.Equal(t, UpgradeRequest{50, 50, p.upgrade}, *u)
		p.CompleteUpgrade()
		p.CompleteRestore()
		require.Equal(t, Capacity(110), p.producerState.maxCapacity)
		require.Equal(t, Capacity(67), p.producerState.capacity)
		require.Panics(t, func() {
			p.CompleteRestore()
		})
		require.Panics(t, func() {
			p.CompleteUpgrade()
		})
	})
}

func TestProducingAgentProduce(t *testing.T) {
	type test struct {
		inProgress  *Booking
		cutOffPrice CapacityUnitPrice
		bids        []Bid
		result      ProductionResult
		state       producerState
	}
	tests := []test{
		{nil, 0, nil, ProductionResult{[]OrderId{}, []OrderId{}}, producerState{100, 100, nil, nil, 0, 0, 0}},
		{nil, 0, []Bid{{100, 20, "1"}}, ProductionResult{[]OrderId{"1"}, []OrderId{}}, producerState{100, 100, nil, nil, 100, 20, 0.2}},
		{nil, 0, []Bid{{40, 20, "1"}, {40, 4, "2"}, {20, 20, "3"}, {40, 1, "4"}}, ProductionResult{[]OrderId{"3", "1", "2"}, []OrderId{"4"}}, producerState{100, 100, nil, nil, 140, 44, 0.1}},
		{nil, 0, []Bid{{100, 100, "1"}, {10, 1, "2"}}, ProductionResult{[]OrderId{"1"}, []OrderId{"2"}}, producerState{100, 100, nil, nil, 110, 100, 1}},
		{nil, 0, []Bid{{200, 20, "1"}}, ProductionResult{[]OrderId{}, []OrderId{}}, producerState{100, 100, nil, &Booking{"1", 100}, 200, 20, 0.1}},
		{&Booking{"1", 100}, 3, nil, ProductionResult{[]OrderId{"1"}, []OrderId{}}, producerState{100, 100, nil, nil, 0, 0, 3}},
		{&Booking{"1", 100}, 3, []Bid{{50, 20, "2"}}, ProductionResult{[]OrderId{"1"}, []OrderId{"2"}}, producerState{100, 100, nil, nil, 50, 0, 3}},
		{&Booking{"1", 50}, 3, []Bid{{100, 20, "2"}}, ProductionResult{[]OrderId{"1"}, []OrderId{}}, producerState{100, 100, nil, &Booking{"2", 50}, 100, 20, 0.2}},
	}
	for idx, tc := range tests {
		// if idx != 3 {
		// 	continue
		// }
		t.Run(fmt.Sprintf("[TEST CASE]: %d\n", idx), func(t *testing.T) {
			p := agent(withInProgres(tc.inProgress), withCutOffPrice(tc.cutOffPrice))
			p.PlaceBids(tc.bids)
			result := p.Produce()
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.state, p.producerState)
		})
	}
}

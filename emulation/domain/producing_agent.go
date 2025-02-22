package domain

import (
	"math"
	"slices"

	"github.com/samber/lo"
)

type ProducerId string

type DegradationRate uint

type Restoration struct {
	Require  ProcessSheet
	Restores Capacity
}

type Upgrade struct {
	Require   ProcessSheet
	Increases Capacity
}

type Booking struct {
	OrderId OrderId
	Booked  Capacity
}

type ProducerInfo struct {
	Id          ProducerId
	PowerType   CapacityType
	Capacity    Capacity
	CutOffPrice CapacityUnitPrice
}

type ProducingAgentConfig struct {
	Id          ProducerId
	Type        CapacityType
	Capacity    Capacity
	Degradation DegradationRate
	Restoration Restoration
	Upgrade     Upgrade
}

type Bid struct {
	Capacity Capacity
	Tokens   Tokens
	OrderId  OrderId
}

type CapacityUnitPrice float32

func (a CapacityUnitPrice) Equal(b CapacityUnitPrice) bool {
	return math.Abs(float64(a-b)) <= 1e-5
}

func (b Bid) CapacityUnitPrice() CapacityUnitPrice {
	return CapacityUnitPrice(float32(b.Tokens) / float32(b.Capacity))
}

func newProducingAgent(config ProducingAgentConfig) *ProducingAgent {
	return &ProducingAgent{
		config.Id, config.Type, config.Degradation, config.Restoration, config.Upgrade,
		producerState{config.Capacity, nil, nil, 0, 0, 0},
	}
}

type producerState struct {
	capacity          Capacity
	bids              []Bid // replace with MaxHeap
	inProgress        *Booking
	requestedCapacity Capacity
	funds             Tokens
	cutOffPrice       CapacityUnitPrice
}

type ProducingAgent struct {
	id          ProducerId
	powerType   CapacityType
	degradation DegradationRate
	restoration Restoration
	upgrade     Upgrade
	state       producerState
}

func (p *ProducingAgent) Info() ProducerInfo {
	return ProducerInfo{p.id, p.powerType, p.state.capacity, p.state.cutOffPrice}
}

func (p *ProducingAgent) powerDegradation() Capacity {
	return Capacity(math.Ceil(float64(p.degradation) * float64(p.state.capacity) / 100))
}

func (p *ProducingAgent) View() ProducingAgentView {
	return ProducingAgentView{p.id, p.state.capacity, p.state.requestedCapacity, p.powerDegradation(), p.upgrade.Increases, p.state.funds}
}

func (p *ProducingAgent) PlaceBids(bids []Bid) {
	p.state.bids = append(p.state.bids, bids...)
}

type ProductionResult struct {
	Completed []OrderId
	Rejected  []OrderId
}

func (p *ProducingAgent) Produce() ProductionResult {
	// sorting slice in descending order by the capacity unit price (most valuable come first)
	slices.SortFunc(p.state.bids, func(a, b Bid) int {
		if a.CapacityUnitPrice() < b.CapacityUnitPrice() {
			return 1
		}
		return -1
	})
	requestedCapacity := lo.SumBy(p.state.bids, func(b Bid) Capacity {
		return b.Capacity
	})
	remainingCapacity := int(p.state.capacity)
	completed := []OrderId{}
	rejected := []OrderId{}
	cutOffPrice := p.state.cutOffPrice
	funds := Tokens(0)
	inProgress := p.state.inProgress
	if inProgress != nil {
		remainingCapacity -= int(inProgress.Booked)
		if remainingCapacity >= 0 {
			completed = append(completed, inProgress.OrderId)
			inProgress = nil
		} else {
			inProgress = &Booking{inProgress.OrderId, inProgress.Booked - p.state.capacity}
		}
	}
	for i := range p.state.bids {
		if remainingCapacity > 0 {
			bid := p.state.bids[i]
			// fmt.Printf("rcap: %d tokens: %d cup: %f\n", bid.Capacity, bid.Tokens, bid.CapacityUnitPrice())
			cutOffPrice = bid.CapacityUnitPrice()
			funds += bid.Tokens
			remainingCapacity -= int(bid.Capacity)
			if remainingCapacity >= 0 {
				completed = append(completed, bid.OrderId)
			} else {
				inProgress = &Booking{bid.OrderId, Capacity(-remainingCapacity)}
			}
			continue
		}
		rejected = append(rejected, p.state.bids[i].OrderId)
	}
	p.state = producerState{p.state.capacity, nil, inProgress, requestedCapacity, funds, cutOffPrice}
	return ProductionResult{completed, rejected}
}

type ProducingAgentView struct {
	Id                ProducerId
	TotalCapacity     Capacity
	RequestedCapacity Capacity
	Degradation       Capacity
	Upgrade           Capacity
	Funds             Tokens
}

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
		producerState{config.Capacity, config.Capacity, nil, nil, 0, 0, 0}, consumerState{},
	}
}

type producerState struct {
	capacity          Capacity
	maxCapacity       Capacity
	bids              []Bid // replace with MaxHeap
	inProgress        *Booking
	requestedCapacity Capacity
	funds             Tokens
	cutOffPrice       CapacityUnitPrice
}

type consumerState struct {
	requestedUpgrades uint
	requestedRestores uint
}

type ProducingAgent struct {
	id            ProducerId
	powerType     CapacityType
	degradation   DegradationRate
	restoration   Restoration
	upgrade       Upgrade
	producerState producerState
	consumerState consumerState
}

func (p *ProducingAgent) Info() ProducerInfo {
	return ProducerInfo{p.id, p.powerType, p.producerState.capacity, p.producerState.cutOffPrice}
}

func (p *ProducingAgent) capacityDegradation() Capacity {
	return Capacity(math.Ceil(float64(p.degradation) * float64(p.producerState.capacity) / 100))
}

func (p *ProducingAgent) View() ProducingAgentView {
	return ProducingAgentView{p.id, p.producerState.capacity, p.producerState.requestedCapacity, p.capacityDegradation(), p.upgrade.Increases, p.producerState.funds}
}

func (p *ProducingAgent) PlaceBids(bids []Bid) {
	p.producerState.bids = append(p.producerState.bids, bids...)
}

type ProductionResult struct {
	Completed []OrderId
	Rejected  []OrderId
}

type Upgrades uint
type Restores uint

type UpgradeRequest struct {
	SelfFunds Tokens
	Loan      Tokens
	Upgrade   Upgrade
}

type RestoreRequest struct {
	SelfFunds   Tokens
	Loan        Tokens
	Restoration Restoration
}

type ProducingAgentCommand struct {
	Loan      Tokens
	DoUpgrade bool
	DoRestore bool
}

func (p *ProducingAgent) Invest(cmd ProducingAgentCommand) (*UpgradeRequest, *RestoreRequest) {
	var upgrade *UpgradeRequest
	var restore *RestoreRequest
	loan := cmd.Loan
	funds := p.producerState.funds
	if cmd.DoUpgrade && cmd.DoRestore {
		loan /= 2
		funds /= 2
	}
	if cmd.DoUpgrade {
		p.consumerState.requestedUpgrades++
		upgrade = &UpgradeRequest{funds, loan, p.upgrade}
	}
	if cmd.DoRestore {
		p.consumerState.requestedRestores++
		restore = &RestoreRequest{funds, loan, p.restoration}
	}
	return upgrade, restore
}

func (p *ProducingAgent) CompleteUpgrade() {
	if p.consumerState.requestedUpgrades == 0 {
		panic("no updgrades requested")
	}
	p.consumerState.requestedUpgrades--
	p.producerState.maxCapacity += p.upgrade.Increases
	p.producerState.capacity += p.upgrade.Increases
}

func (p *ProducingAgent) CompleteRestore() {
	if p.consumerState.requestedRestores == 0 {
		panic("no restores requested")
	}
	p.consumerState.requestedRestores--
	p.producerState.capacity = min(p.producerState.maxCapacity, p.producerState.capacity+p.restoration.Restores)
}

func (p *ProducingAgent) Produce() ProductionResult {
	// sorting bids in descending order by the capacity unit price (most valuable come first)
	slices.SortFunc(p.producerState.bids, func(a, b Bid) int {
		if a.CapacityUnitPrice() < b.CapacityUnitPrice() {
			return 1
		}
		return -1
	})
	requestedCapacity := lo.SumBy(p.producerState.bids, func(b Bid) Capacity {
		return b.Capacity
	})
	remainingCapacity := int(p.producerState.capacity)
	completed := []OrderId{}
	rejected := []OrderId{}
	cutOffPrice := p.producerState.cutOffPrice
	funds := Tokens(0)
	inProgress := p.producerState.inProgress
	if inProgress != nil {
		remainingCapacity -= int(inProgress.Booked)
		if remainingCapacity >= 0 {
			completed = append(completed, inProgress.OrderId)
			inProgress = nil
		} else {
			inProgress = &Booking{inProgress.OrderId, inProgress.Booked - p.producerState.capacity}
		}
	}
	for i := range p.producerState.bids {
		if remainingCapacity > 0 {
			bid := p.producerState.bids[i]
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
		rejected = append(rejected, p.producerState.bids[i].OrderId)
	}
	p.producerState = producerState{p.producerState.capacity, p.producerState.capacity, nil, inProgress, requestedCapacity, funds, cutOffPrice}
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

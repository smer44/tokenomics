package domain

import (
	"slices"

	"github.com/samber/lo"
)

type ProducerId uint

type DegradationRate uint

type Restoration struct {
	Require ProcessSheet
	Power   Power
}

type Upgrade struct {
	Require ProcessSheet
	Power   Power
}

type Task struct {
	OrderId OrderId
	Remains Power
}

type ProducerInfo struct {
	Id          ProducerId
	PowerType   PowerType
	Capacity    Power
	CutOffPrice Tokens
}

type ProducingAgentConfig struct {
	Id          ProducerId
	Type        PowerType
	Capacity    Power
	Degradation DegradationRate
	Restoration Restoration
	Upgrade     Upgrade
}

type Bid struct {
	Power   Power
	Tokens  Tokens
	OrderId OrderId
}

func (b Bid) PowerUnitPrice() Tokens {
	return Tokens(float32(b.Power) / float32(b.Tokens))
}

func newProducingAgent(config ProducingAgentConfig) *ProducingAgent {
	return &ProducingAgent{
		config.Id, config.Type, config.Degradation, config.Restoration, config.Upgrade, config.Capacity, nil, Task{}, nil, 0, 0, 0,
	}
}

type ProducingAgent struct {
	// static data
	id          ProducerId
	powerType   PowerType
	degradation DegradationRate
	restoration Restoration
	upgrade     Upgrade
	// dynamic data
	capacity          Power
	bids              []Bid // replace with MaxHeap
	inProgress        Task
	taken             []Task
	cutOffPrice       Tokens
	requestedCapacity Power
	funds             Tokens
}

func (p *ProducingAgent) Info() ProducerInfo {
	return ProducerInfo{p.id, p.powerType, p.capacity, p.cutOffPrice}
}

func (p *ProducingAgent) powerDegradation() Power {
	return Power(float32(p.capacity) / 100 * float32(p.degradation))
}

func (p *ProducingAgent) View() ProducingAgentView {
	return ProducingAgentView{p.id, p.capacity, p.requestedCapacity, p.powerDegradation(), p.upgrade.Power, p.funds}
}

func (p *ProducingAgent) PlaceBids(bids []Bid) {
	p.bids = append(p.bids, bids...)
}

func (p *ProducingAgent) FixBids() {
	// sorting slice descending by the power unit price
	slices.SortFunc(p.bids, func(a, b Bid) int {
		if a.PowerUnitPrice() > b.PowerUnitPrice() {
			return -1
		}
		if a.PowerUnitPrice() < b.PowerUnitPrice() {
			return 1
		}
		return 0
	})
	p.requestedCapacity = lo.SumBy(p.bids, func(b Bid) Power {
		return b.Power
	})
	remains := int(p.capacity) - int(p.inProgress.Remains)
	p.taken = []Task{}
	for i := 0; remains > 0 && i < len(p.bids); i++ {
		bid := p.bids[i]
		p.taken = append(p.taken, Task{bid.OrderId, bid.Power})
		remains -= int(bid.Power)
	}
	p.bids = nil
}

type ProducingAgentView struct {
	Id                ProducerId
	TotalCapacity     Power
	RequestedCapacity Power
	Degradation       Power
	Upgrade           Power
	Funds             Tokens
}

package domain

import (
	"errors"
	"math"
	"slices"

	"github.com/samber/lo"
)

type ProducerId string

type DegradationRate uint

type Restoration struct {
	Require  Product
	Restores Capacity
}

type Upgrade struct {
	Require   Product
	Increases Capacity
}

type booking struct {
	orderId OrderId
	booked  Capacity
	bid     Bid
}

type ProducerInfo struct {
	Id           ProducerId
	CapacityType CapacityType
	Capacity     Capacity
	CutOffPrice  CapacityUnitPrice
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
	CapacityType CapacityType
	Capacity     Capacity
	Tokens       Tokens
	OrderId      OrderId
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
	inProgress        *booking
	requestedCapacity Capacity
	funds             Tokens
	cutOffPrice       CapacityUnitPrice
}

type RestorationRunning bool
type UpgradeRunning bool

type consumerState struct {
	upgradeRunning     UpgradeRunning
	restorationRunning RestorationRunning
}

type ProducingAgent struct {
	id            ProducerId
	capacityType  CapacityType
	degradation   DegradationRate
	restoration   Restoration
	upgrade       Upgrade
	producerState producerState
	consumerState consumerState
}

func (p *ProducingAgent) Info() ProducerInfo {
	return ProducerInfo{p.id, p.capacityType, p.producerState.capacity, p.producerState.cutOffPrice}
}

func (p *ProducingAgent) capacityDegradation() Capacity {
	return Capacity(math.Ceil(float64(p.degradation) * float64(p.producerState.capacity) / 100))
}

func (p *ProducingAgent) View() ProducingAgentView {
	return ProducingAgentView{p.id, p.producerState.maxCapacity, p.producerState.capacity, p.producerState.requestedCapacity, p.capacityDegradation(), p.upgrade.Increases, p.restoration.Restores, p.consumerState.upgradeRunning, p.consumerState.restorationRunning}
}

func (p *ProducingAgent) PlaceBids(bids []Bid) {
	for i := range bids {
		if bids[i].CapacityType != p.capacityType {
			panic("wrong capacity type")
		}
	}
	p.producerState.bids = append(p.producerState.bids, bids...)
}

type ProductionResult struct {
	Processing []Bid
	Completed  []Bid
	Rejected   []Bid
}

type ProducingAgentCommand struct {
	DoRestoration bool
	DoUpgrade     bool
}

type InvestmentType byte

const (
	InvestmentTypeUpgrade InvestmentType = iota
	InvestmentTypeRestoration
)

type InvestmentRequest struct {
	ProducerId  ProducerId
	Type        InvestmentType
	Product     Product
	CutOffPrice CapacityUnitPrice
}

func (p *ProducingAgent) Invest(cmd ProducingAgentCommand) ([]InvestmentRequest, error) {
	if bool(p.consumerState.upgradeRunning) && cmd.DoUpgrade {
		return nil, errors.New("upgrade is running")
	}
	if bool(p.consumerState.restorationRunning) && cmd.DoRestoration {
		return nil, errors.New("restorations is running")
	}
	requests := []InvestmentRequest{}
	if cmd.DoUpgrade {
		requests = append(requests, InvestmentRequest{p.id, InvestmentTypeUpgrade, p.upgrade.Require, p.producerState.cutOffPrice})
		p.consumerState.upgradeRunning = true
	}
	if cmd.DoRestoration {
		requests = append(requests, InvestmentRequest{p.id, InvestmentTypeRestoration, p.restoration.Require, p.producerState.cutOffPrice})
		p.consumerState.restorationRunning = true
	}
	return requests, nil
}

var ErrNoUpgradesRunning = errors.New("no updgrades running")
var ErrNoRestorationRunning = errors.New("no restoration running")

func (p *ProducingAgent) InvesetmentCompleted(request *InvestmentRequest) {
	if request.ProducerId != p.id {
		panic(ErrNotFound)
	}
	switch request.Type {
	case InvestmentTypeRestoration:
		if !p.consumerState.restorationRunning {
			panic(ErrNoRestorationRunning)
		}
		p.consumerState.restorationRunning = false
		p.producerState.capacity = min(p.producerState.maxCapacity, p.producerState.capacity+p.restoration.Restores)
	case InvestmentTypeUpgrade:
		if !p.consumerState.upgradeRunning {
			panic(ErrNoUpgradesRunning)
		}
		p.consumerState.upgradeRunning = false
		p.producerState.maxCapacity += p.upgrade.Increases
		p.producerState.capacity += p.upgrade.Increases

	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *ProducingAgent) InvesetmentRejected(request *InvestmentRequest) {
	if request.ProducerId != p.id {
		panic(ErrNotFound)
	}
	switch request.Type {
	case InvestmentTypeRestoration:
		if !p.consumerState.restorationRunning {
			panic(ErrNoRestorationRunning)
		}
		p.consumerState.restorationRunning = false
	case InvestmentTypeUpgrade:
		if !p.consumerState.upgradeRunning {
			panic(ErrNoUpgradesRunning)
		}
		p.consumerState.upgradeRunning = false
	default:
		panic(errors.ErrUnsupported)
	}
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
	completed := []Bid{}
	rejected := []Bid{}
	processing := []Bid{}
	cutOffPrice := p.producerState.cutOffPrice
	funds := Tokens(0)
	inProgress := p.producerState.inProgress
	if inProgress != nil {
		remainingCapacity -= int(inProgress.booked)
		if remainingCapacity >= 0 {
			completed = append(completed, inProgress.bid)
			inProgress = nil
		} else {
			inProgress = &booking{inProgress.orderId, inProgress.booked - p.producerState.capacity, inProgress.bid}
		}
	}
	for i := range p.producerState.bids {
		if remainingCapacity > 0 {
			bid := p.producerState.bids[i]
			cutOffPrice = bid.CapacityUnitPrice()
			funds += bid.Tokens
			remainingCapacity -= int(bid.Capacity)
			if remainingCapacity >= 0 {
				completed = append(completed, bid)
			} else {
				inProgress = &booking{bid.OrderId, Capacity(-remainingCapacity), bid}
			}
			continue
		}
		rejected = append(rejected, p.producerState.bids[i])
	}
	p.producerState = producerState{p.producerState.capacity, p.producerState.capacity, nil, inProgress, requestedCapacity, funds, cutOffPrice}
	if inProgress != nil {
		processing = append(processing, inProgress.bid)
	}
	return ProductionResult{processing, completed, rejected}
}

type ProducingAgentView struct {
	Id                 ProducerId
	MaxCapacity        Capacity
	Capacity           Capacity
	RequestedCapacity  Capacity
	Degradation        Capacity
	Upgrade            Capacity
	Restoration        Capacity
	UpgradeRunning     UpgradeRunning
	RestorationRunning RestorationRunning
}

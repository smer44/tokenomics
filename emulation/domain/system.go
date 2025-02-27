package domain

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
)

var ErrNotFound = errors.New("not found")
var ErrWrongState = errors.New("wrong state")

type SystemState byte

const (
	SystemStateOrdersPlacement SystemState = iota
	SystemStateOrdering
)

type OrderingAgentId string

func FromProducerId(id ProducerId) OrderingAgentId {
	return OrderingAgentId(fmt.Sprintf("p%s", id))
}
func FromConsumerId(id ConsumerId) OrderingAgentId {
	return OrderingAgentId(fmt.Sprintf("c%s", id))
}

type System struct {
	state           SystemState
	idGen           OrderIdGenerator
	cycleEmission   Tokens
	investmentFund  Tokens
	processSheets   map[Product]ProcessSheet
	producerLookup  map[CapacityType][]ProducerId
	producingAgents map[ProducerId]*ProducingAgent
	producerInfos   map[ProducerId]ProducerInfo
	orderingAgents  map[OrderingAgentId]*OrderingAgent
	orders          map[OrderId]*Order
	consumers       map[ConsumerId]Consumer
}

func NewSystem(idGen OrderIdGenerator, emission Tokens, ps []ProcessSheet, pa []ProducingAgentConfig, consumers map[ConsumerId]Consumer) *System {
	s := &System{
		SystemStateOrdersPlacement,
		idGen,
		emission,
		0,
		lo.SliceToMap(ps, func(ps ProcessSheet) (Product, ProcessSheet) {
			return ps.Product, ps
		}),
		lo.SliceToMap(pa, func(p ProducingAgentConfig) (CapacityType, []ProducerId) {
			return p.Type, []ProducerId{p.Id}
		}),
		lo.SliceToMap(pa, func(p ProducingAgentConfig) (ProducerId, *ProducingAgent) {
			return p.Id, newProducingAgent(p)
		}),
		nil,
		map[OrderingAgentId]*OrderingAgent{},
		map[OrderId]*Order{},
		consumers,
	}
	s.producerInfos = lo.MapEntries(s.producingAgents, func(id ProducerId, ps *ProducingAgent) (ProducerId, ProducerInfo) {
		return id, ps.Info()
	})
	for _, c := range consumers {
		id := FromConsumerId(c.Id())
		s.orderingAgents[id] = NewOrderingAgent(id)
	}
	for prodId := range s.producingAgents {
		id := FromProducerId(prodId)
		s.orderingAgents[id] = NewOrderingAgent(id)
	}
	s.startCycle()
	return s
}

func (s *System) OrderingAgentView(id OrderingAgentId) (OrderingAgentView, error) {
	if s.state != SystemStateOrdering {
		return OrderingAgentView{}, ErrWrongState
	}
	agent, ok := s.orderingAgents[id]
	if !ok {
		return OrderingAgentView{}, ErrNotFound
	}
	return agent.View(s.producerInfos), nil
}

func (s *System) OrderingAgentAction(id OrderingAgentId, cmd OrderingAgentCommand) error {
	if s.state != SystemStateOrdering {
		return ErrWrongState
	}
	agent, ok := s.orderingAgents[id]
	if !ok {
		return ErrNotFound
	}
	bids, err := agent.Bidding(cmd, s.producerInfos)
	if err != nil {
		return err
	}
	for prodId, bids := range bids {
		p, ok := s.producingAgents[prodId]
		if !ok {
			return ErrNotFound
		}
		p.PlaceBids(bids)
	}
	return nil
}

func (s *System) ProducingAgentView(id ProducerId) (ProducingAgentView, error) {
	if s.state != SystemStateOrdersPlacement {
		return ProducingAgentView{}, ErrWrongState
	}
	p, ok := s.producingAgents[id]
	if !ok {
		return ProducingAgentView{}, ErrNotFound
	}
	return p.View(), nil
}

func (s *System) ProducingAgentAction(id ProducerId, cmd ProducingAgentCommand) error {
	if s.state != SystemStateOrdersPlacement {
		return ErrWrongState
	}
	p, ok := s.producingAgents[id]
	if !ok {
		return ErrNotFound
	}
	investmentRequests, err := p.Invest(cmd)
	if err != nil {
		return err
	}
	for _, r := range investmentRequests {
		id := s.idGen.New()
		s.orders[id] = NewInvestmentOrder(id, MustGet(s.processSheets, r.Product), r)
	}
	return nil
}

func (s *System) placeComsumersOrders() {
	for _, c := range s.consumers {
		for _, request := range c.Order() {
			id := s.idGen.New()
			order := NewConsumerOrder(id, MustGet(s.processSheets, request.Product), request)
			s.orders[id] = order
		}
	}
}

func (s *System) emit() {
	s.investmentFund = s.cycleEmission / 2
	consumerTokens := s.cycleEmission / 2 / Tokens(len(s.consumers))
	for _, c := range s.consumers {
		c.Emit(consumerTokens)
	}
}

func (s *System) startCycle() {
	s.state = SystemStateOrdersPlacement
	s.emit()
	s.placeComsumersOrders()
}

func (s *System) StartOrdering() error {
	if s.state != SystemStateOrdersPlacement {
		return ErrWrongState
	}
	// Make funding
	// Distibute investment fund accodingly producer's cut off price  (capacity deficit)
	totalCutOffSum := CapacityUnitPrice(0)
	requireFunding := map[OrderId]*Order{}
	for orderId, order := range s.orders {
		if !order.RequiresFunding() {
			continue
		}
		requireFunding[orderId] = order
		totalCutOffSum += order.CutOffPrice()
	}
	for _, order := range requireFunding {
		funds := Tokens(float32(order.CutOffPrice()) * float32(s.investmentFund) / float32(totalCutOffSum))
		order.Fund(funds)
	}
	// Place orders
	for _, order := range s.orders {
		MustGet(s.orderingAgents, order.AgentId()).PlaceOrder(order.Info())
	}
	s.state = SystemStateOrdering
	return nil
}

type CycleResult struct {
	Score Score
}

func (s *System) EndCycle() (CycleResult, error) {
	if s.state != SystemStateOrdering {
		return CycleResult{}, ErrWrongState
	}
	for _, p := range s.producingAgents {
		result := p.Produce()
		for _, bid := range result.Processing {
			MustGet(s.orders, bid.OrderId).Processing(bid.CapacityType, bid.Tokens)
		}
		for _, bid := range result.Completed {
			MustGet(s.orders, bid.OrderId).Completed(bid.CapacityType, bid.Tokens)
		}
		for _, bid := range result.Rejected {
			MustGet(s.orders, bid.OrderId).Rejected(bid.CapacityType)
		}
	}
	cycleScore := Score(0)
	for id, order := range s.orders {
		score, event := order.EndCycle()
		cycleScore += score
		completed := true
		switch e := event.(type) {
		case ConsumerRequestCompleted:
		case InvestmentRequestCompleted:
			s.producingAgents[e.Request.ProducerId].InvesetmentCompleted(e.Request)
		case ConsumerRequestRejected:
			s.consumers[e.Request.ConsumerId].Emit(e.Remaining)
		case InvestmentRequestRejected:
			s.producingAgents[e.Request.ProducerId].InvesetmentRejected(e.Request)
		case OrderStillProcessing:
			completed = false
		default:
			panic(errors.ErrUnsupported)
		}
		if completed {
			delete(s.orders, id)
		}
	}
	s.startCycle()
	return CycleResult{cycleScore}, nil
}

type ProcessSheet struct {
	Product Product                   `json:"product"`
	Require map[CapacityType]Capacity `json:"require"`
}

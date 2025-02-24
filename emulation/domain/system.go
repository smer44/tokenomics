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
	SystemStateOrderPlacement = iota
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
	emission        Tokens
	investmentFunds Tokens
	processSheets   map[Product]ProcessSheet
	powerProviders  map[CapacityType][]ProducerId // single procuder for each type at the moment
	producers       map[ProducerId]*ProducingAgent
	producerInfos   map[ProducerId]ProducerInfo
	orderingAgents  map[OrderingAgentId]*OrderingAgent
	orders          map[OrderId]*Order
	consumers       []Consumer
}

func NewSystem(idGen OrderIdGenerator, emission Tokens, ps []ProcessSheet, pa []ProducingAgentConfig, consumers []Consumer) *System {
	s := &System{
		SystemStateOrderPlacement,
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
	s.producerInfos = lo.MapEntries(s.producers, func(id ProducerId, ps *ProducingAgent) (ProducerId, ProducerInfo) {
		return id, ps.Info()
	})
	for _, c := range consumers {
		id := FromConsumerId(c.Id())
		s.orderingAgents[id] = NewOrderingAgent(id)
	}
	for prodId := range s.producers {
		id := FromProducerId(prodId)
		s.orderingAgents[id] = NewOrderingAgent(id)
	}
	s.startTurn()

	return s
}

func (s *System) emit() {
	s.investmentFunds = s.emission / 2
	consumerTokens := s.emission / 2 / Tokens(len(s.orderingAgents))
	for _, c := range s.consumers {
		c.Emit(consumerTokens)
	}
}

func (s *System) getProcessSheet(id Product) ProcessSheet {
	ps, ok := s.processSheets[id]
	if !ok {
		panic(fmt.Sprintf("no process sheet for [%d]", id))
	}
	return ps
}

func (s *System) placeComsumersOrders() {
	for _, c := range s.consumers {
		for _, request := range c.Order() {
			id := s.idGen.New()
			order := newCustomerOrder(id, request.Tokens, s.getProcessSheet(request.Product), request)
			s.orders[id] = order
		}
	}
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
	bids, err := agent.Bidding(s.producerInfos, cmd)
	if err != nil {
		return err
	}
	for prodId, bids := range bids {
		p, ok := s.producers[prodId]
		if !ok {
			return ErrNotFound
		}
		p.PlaceBids(bids)
	}
	return nil
}

func (s *System) ProducingAgentView(id ProducerId) (ProducingAgentView, error) {
	if s.state != SystemStateOrderPlacement {
		return ProducingAgentView{}, ErrWrongState
	}
	p, ok := s.producers[id]
	if !ok {
		return ProducingAgentView{}, ErrNotFound
	}
	return p.View(), nil
}

func (s *System) ProducingAgentAction(id ProducerId, cmd ProducingAgentCommand) error {
	if s.state != SystemStateOrderPlacement {
		return ErrWrongState
	}
	p, ok := s.producers[id]
	if !ok {
		return ErrNotFound
	}
	investmentRequests, err := p.Invest(cmd)
	if err != nil {
		return err
	}
	for _, r := range investmentRequests {
		id := s.idGen.New()
		s.orders[id] = newInvestmentOrder(id, s.getProcessSheet(r.Product), r)
	}
	return nil
}

func (s *System) startTurn() {
	s.state = SystemStateOrderPlacement
	s.emit()
	s.placeComsumersOrders()
}

type Score uint

type TurnResult struct {
	Score Score
}

func (s *System) StartOrdering() error {
	if s.state != SystemStateOrderPlacement {
		return ErrWrongState
	}
	// Make funding
	// Distibute investment fund accodingly producer's cut off price  (capacity deficit)
	totalCutOffSum := CapacityUnitPrice(0)
	funding := map[OrderId]*Order{}
	for orderId, order := range s.orders {
		if order.Status() != OrderStatusFunding {
			continue
		}
		funding[orderId] = order
		totalCutOffSum += order.CutOffPrice()
	}
	for _, order := range funding {
		funds := Tokens(float32(order.CutOffPrice()) * float32(s.investmentFunds) / float32(totalCutOffSum))
		order.Fund(funds)
	}
	s.investmentFunds = 0
	// Place orders
	for _, order := range s.orders {
		if order.Status() != OrderStatusPlacing {
			continue
		}
		id := order.AgentId()
		agent, ok := s.orderingAgents[id]
		if !ok {
			panic(fmt.Sprintf("no ordering agent with id [%s]", id))
		}
		agent.PlaceOrder(order.Info())
	}
	s.state = SystemStateOrdering
	return nil
}

func (s *System) EndTurn() (TurnResult, error) {
	if s.state != SystemStateOrdering {
		return TurnResult{}, ErrWrongState
	}
	for _, p := range s.producers {
		result := p.Produce()
		for _, bid := range result.Processing {
			ord, ok := s.orders[bid.OrderId]
			if !ok {
				panic(ErrNotFound)
			}
			ord.Processing(bid)
		}
	}
	s.startTurn()
	return TurnResult{}, nil
}

type ProcessSheet struct {
	Product Product                   `json:"product"`
	Require map[CapacityType]Capacity `json:"require"`
}

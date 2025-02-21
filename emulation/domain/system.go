package domain

import (
	"errors"
	"fmt"

	"github.com/samber/lo"
)

var ErrNotFound = errors.New("not found")
var ErrWrongState = errors.New("wrong state")

type SystemState uint

const (
	SystemStateOrdering SystemState = iota
	SystemStateProducing
)

type System struct {
	idGen           OrderIdGenerator
	state           SystemState
	emission        Tokens
	investementPool Tokens
	processSheets   map[Product]ProcessSheet
	powerProviders  map[PowerType][]ProducerId // single procuder for each type at the moment
	producers       map[ProducerId]*ProducingAgent
	orderingAgents  map[ConsumerId]*OrderingAgent
}

func NewSystem(idGen OrderIdGenerator, emission Tokens, ps []ProcessSheet, pa []ProducingAgentConfig, c []Consumer) *System {
	s := &System{
		idGen,
		SystemStateOrdering,
		emission,
		0,
		lo.SliceToMap(ps, func(ps ProcessSheet) (Product, ProcessSheet) {
			return ps.Product, ps
		}),
		// DO FOR MULTIPLE PRODUCERS WITH SAME POWER TYPE
		lo.SliceToMap(pa, func(p ProducingAgentConfig) (PowerType, []ProducerId) {
			return p.Type, []ProducerId{p.Id}
		}),
		lo.SliceToMap(pa, func(p ProducingAgentConfig) (ProducerId, *ProducingAgent) {
			return p.Id, newProducingAgent(p)
		}),
		lo.SliceToMap(c, func(c Consumer) (ConsumerId, *OrderingAgent) {
			return c.Id(), &OrderingAgent{[]*Order{}, map[OrderId]*Order{}, c}
		}),
	}
	s.startTurn()
	return s
}

func (s *System) emit() {
	s.investementPool = s.emission / 2
	consumerTokens := s.emission / 2 / Tokens(len(s.orderingAgents))
	for _, oa := range s.orderingAgents {
		oa.consumer.Emit(consumerTokens)
	}
}

func (s *System) placeOrders() {
	for _, oa := range s.orderingAgents {
		for _, request := range oa.consumer.Order() {
			oa.PlaceOrder(s.idGen.New(), request, s.processSheets)
		}
	}
}

func (s *System) OrderingAgentView(id ConsumerId) (OrderingAgentView, error) {
	if s.state != SystemStateOrdering {
		return OrderingAgentView{}, ErrWrongState
	}
	agent, ok := s.orderingAgents[id]
	if !ok {
		return OrderingAgentView{}, ErrNotFound
	}
	return agent.View(s.producers), nil
}

func (s *System) OrderingAgentAction(id ConsumerId, cmd OrderingAgentCommand) error {
	if s.state != SystemStateOrdering {
		return ErrWrongState
	}
	for p, bids := range cmd.Bids {
		producer, ok := s.producers[p]
		if !ok {
			return fmt.Errorf("%w: producer [%d]", ErrNotFound, p)
		}
		producer.PlaceBids(bids)
	}
	return nil
}

func (s *System) FixBids() error {
	if s.state != SystemStateOrdering {
		return ErrWrongState
	}
	for _, p := range s.producers {
		p.FixBids()
	}
	s.state = SystemStateProducing
	return nil
}

type ProducingAgentCommand struct {
}

func (s *System) ProducingAgentView(id ProducerId) (ProducingAgentView, error) {
	if s.state != SystemStateProducing {
		return ProducingAgentView{}, ErrWrongState
	}
	p, ok := s.producers[id]
	if !ok {
		return ProducingAgentView{}, ErrNotFound
	}
	return p.View(), nil
}

func (s *System) ProducingAgentAction(id ProducerId, cmd ProducingAgentCommand) error {
	return nil
}

func (s *System) startTurn() {
	s.emit()
	s.placeOrders()
	s.state = SystemStateOrdering
}

type TurnResult struct {
}

func (s *System) EndTurn() (TurnResult, error) {
	s.startTurn()
	return TurnResult{}, nil
}

type ProcessSheet struct {
	Product Product             `json:"product"`
	Require map[PowerType]Power `json:"require"`
}

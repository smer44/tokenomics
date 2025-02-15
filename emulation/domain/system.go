package domain

import (
	"errors"

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
	producingAgents map[PowerType]*ProducingAgent // single procuder for each type at the moment
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
		lo.SliceToMap(pa, func(p ProducingAgentConfig) (PowerType, *ProducingAgent) {
			return p.Type, &ProducingAgent{p.Id, p.Capacity, p.Degradation, p.Restoration, p.Upgrade, nil, nil}
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
	return agent.View(), nil
}

func (s *System) OrderingAgentAction(id ConsumerId, cmd OrderingAgentCommand) error {
	return nil
}

func (s *System) FixBids() {
}

type ProducingAgentView struct {
}

type ProducingAgentCommand struct {
}

func (s *System) ProducingAgentView(id ProducerId) (ProducingAgentView, error) {
	return ProducingAgentView{}, nil
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

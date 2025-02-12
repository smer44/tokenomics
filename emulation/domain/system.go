package domain

import (
	"fmt"

	"github.com/samber/lo"
)

type Tokens uint
type Product uint
type PowerType uint
type Consumer uint
type DegradationRate uint
type Power struct {
	t     PowerType
	value uint
}
type OrderId string

func (p Power) Comparable(a Power) bool {
	return p.t == a.t
}

func (p Power) Equal(a Power) bool {
	if !p.Comparable(a) {
		panic(fmt.Sprintf("product [%d] can't be compared with [%d]", p.t, a.t))
	}
	return p.value == a.value
}

type System struct {
	investementPool Tokens
	processSheets   map[Product]ProcessSheet
	producingAgents map[Product]ProducingAgentImpl // single procuder for each type at the moment
	orderingAgents  map[Consumer]OrderingAgentImpl
}

func NewSystem(ps []ProcessSheet, pa []ProducingAgentImpl, c []Consumer) *System {
	return &System{
		0,
		lo.SliceToMap(ps, func(ps ProcessSheet) (Product, ProcessSheet) {
			return ps.Product, ps
		}),
		lo.SliceToMap(pa, func(pa ProducingAgentImpl) (Product, ProducingAgentImpl) {
			return pa.product, pa
		}),
		lo.SliceToMap(c, func(c Consumer) (Consumer, OrderingAgentImpl) {
			return c, OrderingAgentImpl{[]*Order{}, map[OrderId]*Order{}}
		}),
	}
}

type ProcessSheet struct {
	Product Product           `json:"product"`
	Parts   map[Product]Power `json:"parts"`
}

type OrderingAgentImpl struct {
	incoming   []*Order
	inProgress map[OrderId]*Order
}

type OrderingAgentState struct {
	Incoming   []*Order
	InProgress []*Order
}

func (a *OrderingAgentImpl) State() OrderingAgentState {
	return OrderingAgentState{a.incoming, lo.Values(a.inProgress)}
}

type Restoration struct {
	Required ProcessSheet
	Power    Power
}

type Upgrade struct {
	Required ProcessSheet
	Power    Power
}

type Task struct {
	OrderId OrderId
	Agent   *OrderingAgentImpl
}

type Bid struct {
	Power   Power
	Tokens  Tokens
	OrderId OrderId
	Agent   *OrderingAgentImpl
}

type ProducingAgentImpl struct {
	product     Product
	capacity    Power
	degradation DegradationRate
	bids        []Bid // replace with MaxHeap
	restoration Restoration
	upgrade     Upgrade
	inProgress  []Task
}

func NewProducingAgent(product Product, capacity Power, degradation DegradationRate, restoration Restoration, upgrade Upgrade) *ProducingAgentImpl {
	return &ProducingAgentImpl{product, capacity, degradation, []Bid{}, restoration, upgrade, []Task{}}
}

type Part struct {
	Required Power
	Ready    Power
}

type Order struct {
	Id     OrderId
	Tokens Tokens
	Parts  map[Product]Part
}

func NewOrder(id OrderId, t Tokens, ps *ProcessSheet) *Order {
	if t <= 0 {
		panic("tokens must be greater than 0")
	}
	parts := make(map[Product]Part, len(ps.Parts))
	for p, v := range ps.Parts {
		parts[p] = Part{v, Power{p, 0}}
	}
	return &Order{id, t, parts}
}

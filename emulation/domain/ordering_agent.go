package domain

import (
	"fmt"

	"github.com/samber/lo"
)

type OrderId string

type CapacityValue struct {
	Type  CapacityType
	Power Capacity
}

type Request struct {
	Tokens     Tokens
	Capacities []CapacityValue
}

type OrderingAgentView struct {
	Incoming   map[OrderId]Request
	InProgress map[OrderId]Request
	Producers  map[CapacityType][]ProducerInfo
}

type OrderingAgentCommand struct {
	Bids map[ProducerId][]Bid
}

type OrderingAgent struct {
	incoming   []*Order
	inProgress map[OrderId]*Order
	consumer   Consumer
}

func (oa *OrderingAgent) PlaceOrder(id OrderId, r ProductRequest, pss map[Product]ProcessSheet) {
	ps, ok := pss[r.Product]
	if !ok {
		panic(fmt.Sprintf("not found product [%d]", r.Product))
	}
	oa.incoming = append(oa.incoming, newOrder(id, r.Tokens, &ps))
}

func (oa *OrderingAgent) View(producers map[ProducerId]*ProducingAgent) OrderingAgentView {
	powers := map[CapacityType]struct{}{}
	result := OrderingAgentView{
		Incoming: lo.SliceToMap(oa.incoming, func(o *Order) (OrderId, Request) {
			return o.Id, Request{
				Tokens: o.Tokens,
				Capacities: lo.MapToSlice(o.Parts, func(pt CapacityType, part Part) CapacityValue {
					powers[pt] = struct{}{}
					return CapacityValue{pt, part.Remains()}
				}),
			}
		}),
		InProgress: lo.SliceToMap(lo.Values(oa.inProgress), func(o *Order) (OrderId, Request) {
			return o.Id, Request{
				Tokens: o.Tokens,
				Capacities: lo.MapToSlice(o.Parts, func(pt CapacityType, part Part) CapacityValue {
					powers[pt] = struct{}{}
					return CapacityValue{pt, part.Remains()}
				}),
			}
		}),
	}
	result.Producers = make(map[CapacityType][]ProducerInfo, len(powers))
	for _, p := range producers {
		_, ok := powers[p.powerType]
		if !ok {
			continue
		}
		result.Producers[p.powerType] = append(result.Producers[p.powerType], p.Info())
	}
	return result
}

type Part struct {
	Required Capacity
	Consumed Capacity
}

func (p Part) Remains() Capacity {
	return p.Required - p.Consumed
}

type Order struct {
	Id     OrderId
	Tokens Tokens
	Parts  map[CapacityType]Part
}

type OrderIdGenerator interface {
	New() OrderId
}

func newOrder(id OrderId, t Tokens, ps *ProcessSheet) *Order {
	if t <= 0 {
		panic("tokens must be greater than 0")
	}
	parts := make(map[CapacityType]Part, len(ps.Require))
	for p, v := range ps.Require {
		parts[p] = Part{v, 0}
	}
	return &Order{id, t, parts}
}

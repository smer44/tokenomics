package domain

import (
	"fmt"

	"github.com/samber/lo"
)

type OrderId string

type PowerValue struct {
	Type  PowerType
	Power Power
}

type Request struct {
	Tokens Tokens
	Powers []PowerValue
}

type OrderingAgentView struct {
	Incoming   map[OrderId]Request
	InProgress map[OrderId]Request
	Producers  map[ProducerId]ProducerInfo
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

func (oa *OrderingAgent) View() OrderingAgentView {
	return OrderingAgentView{
		Incoming: lo.SliceToMap(oa.incoming, func(o *Order) (OrderId, Request) {
			return o.Id, Request{
				Tokens: o.Tokens,
				Powers: lo.MapToSlice(o.Parts, func(pt PowerType, part Part) PowerValue {
					return PowerValue{pt, part.Remains()}
				}),
			}
		}),
		InProgress: lo.SliceToMap(lo.Values(oa.inProgress), func(o *Order) (OrderId, Request) {
			return o.Id, Request{
				Tokens: o.Tokens,
				Powers: lo.MapToSlice(o.Parts, func(pt PowerType, part Part) PowerValue {
					return PowerValue{pt, part.Remains()}
				}),
			}
		}),
	}
}

type Part struct {
	Required Power
	Consumed Power
}

func (p Part) Remains() Power {
	return p.Required - p.Consumed
}

type Order struct {
	Id     OrderId
	Tokens Tokens
	Parts  map[PowerType]Part
}

type OrderIdGenerator interface {
	New() OrderId
}

func newOrder(id OrderId, t Tokens, ps *ProcessSheet) *Order {
	if t <= 0 {
		panic("tokens must be greater than 0")
	}
	parts := make(map[PowerType]Part, len(ps.Require))
	for p, v := range ps.Require {
		parts[p] = Part{v, 0}
	}
	return &Order{id, t, parts}
}

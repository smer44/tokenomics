package domain

import "fmt"

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
	Incoming   map[OrderId]map[CapacityType]Capacity
	InProgress map[OrderId]map[CapacityType]Capacity
	Producers  map[CapacityType]map[ProducerId]ProducerInfo
}

type OrderingAgentCommand struct {
	Bids map[OrderId]map[ProducerId]Tokens
}

type OrderingAgent struct {
	id         OrderingAgentId
	incoming   map[OrderId]OrderInfo
	inProgress map[OrderId]OrderInfo
}

func (oa *OrderingAgent) PlaceOrder(orderInfo OrderInfo) {
	oa.incoming[orderInfo.Id] = orderInfo
}

func (oa *OrderingAgent) View(producers map[ProducerId]ProducerInfo) OrderingAgentView {
	// powers := map[CapacityType]struct{}{}
	// result := OrderingAgentView{
	// 	Incoming: lo.SliceToMap(oa.incoming, func(o *Order) (OrderId, Request) {
	// 		return o.Id, Request{
	// 			Tokens: o.Tokens,
	// 			Capacities: lo.MapToSlice(o.Parts, func(pt CapacityType, part Part) CapacityValue {
	// 				powers[pt] = struct{}{}
	// 				return CapacityValue{pt, part.Remains()}
	// 			}),
	// 		}
	// 	}),
	// 	InProgress: lo.SliceToMap(lo.Values(oa.inProgress), func(o *Order) (OrderId, Request) {
	// 		return o.Id, Request{
	// 			Tokens: o.Tokens,
	// 			Capacities: lo.MapToSlice(o.Parts, func(pt CapacityType, part Part) CapacityValue {
	// 				powers[pt] = struct{}{}
	// 				return CapacityValue{pt, part.Remains()}
	// 			}),
	// 		}
	// 	}),
	// }
	// result.Producers = make(map[CapacityType][]ProducerInfo, len(powers))
	// for _, p := range producers {
	// 	_, ok := powers[p.powerType]
	// 	if !ok {
	// 		continue
	// 	}
	// 	result.Producers[p.powerType] = append(result.Producers[p.powerType], p.Info())
	// }
	// return result
	return OrderingAgentView{}
}

func (oa *OrderingAgent) Bidding(producers map[ProducerId]ProducerInfo, cmd OrderingAgentCommand) (map[ProducerId][]Bid, error) {
	for orderId, p := range cmd.Bids {
		order, ok := oa.incoming[orderId]
		if !ok {
			return nil, fmt.Errorf("%w: order id [%s] not found for agent [%s]", ErrNotFound, orderId, oa.id)
		}
		positions := 0
		for prodId, bid := range p {

		}
	}
}

func NewOrderingAgent(id OrderingAgentId) *OrderingAgent {
	return &OrderingAgent{id, nil, map[OrderId]OrderInfo{}}
}

package domain

import (
	"fmt"
	"log/slog"

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
	Incoming  map[OrderId]map[CapacityType]Capacity
	Producers map[CapacityType]map[ProducerId]ProducerInfo
}

type OrderingAgentCommand struct {
	Orders map[OrderId]map[ProducerId]Tokens
}

type OrderingAgent struct {
	id         OrderingAgentId
	incoming   map[OrderId]OrderInfo
	cmdHandled bool
}

func (oa *OrderingAgent) PlaceOrder(orderInfo OrderInfo) {
	if orderInfo.Fulfilled() {
		return
	}
	oa.incoming[orderInfo.Id] = orderInfo
	logEvent("ordering.order.placed",
		withOrderId(orderInfo.Id),
		withTokens(orderInfo.Tokens))
}

func (oa *OrderingAgent) View(producers map[ProducerId]ProducerInfo) OrderingAgentView {
	capacityTypes := map[CapacityType]struct{}{}
	result := OrderingAgentView{
		Incoming: lo.MapEntries(oa.incoming, func(id OrderId, oi OrderInfo) (OrderId, map[CapacityType]Capacity) {
			for ct := range oi.Required {
				capacityTypes[ct] = struct{}{}
			}
			return id, oi.Required
		}),
	}
	result.Producers = make(map[CapacityType]map[ProducerId]ProducerInfo, len(capacityTypes))
	for produerId, p := range producers {
		_, ok := capacityTypes[p.CapacityType]
		if !ok {
			continue
		}
		producers, ok := result.Producers[p.CapacityType]
		if !ok {
			producers = map[ProducerId]ProducerInfo{}
		}
		producers[produerId] = p
		result.Producers[p.CapacityType] = producers
	}
	return result
}

func (oa *OrderingAgent) CompleteCycle() {
	logEvent("ordering.cycle.completed",
		slog.String("agentId", string(oa.id)),
		slog.Int("orders", len(oa.incoming)))
	oa.incoming, oa.cmdHandled = map[OrderId]OrderInfo{}, true
}

func (oa *OrderingAgent) HandleCmd(cmd OrderingAgentCommand, producers map[ProducerId]ProducerInfo) (map[ProducerId][]Bid, error) {
	if oa.cmdHandled {
		return nil, fmt.Errorf("command is handled already")
	}
	if len(cmd.Orders) != len(oa.incoming) {
		return nil, fmt.Errorf("too few orders passed. Incoming [%d] passed [%d]", len(oa.incoming), len(cmd.Orders))
	}

	logEvent("ordering.command.received",
		slog.String("agentId", string(oa.id)),
		slog.Int("orders", len(cmd.Orders)))

	result := map[ProducerId][]Bid{}
	for orderId, bids := range cmd.Orders {
		order, ok := oa.incoming[orderId]
		if !ok {
			return nil, fmt.Errorf("%w: order id [%s] not found for agent [%s]", ErrNotFound, orderId, oa.id)
		}
		if len(order.Required) != len(bids) {
			return nil, fmt.Errorf("too few bids passed for order [%s]", orderId)
		}
		agentBids := lo.Sum(lo.Values(bids))
		if order.Tokens != agentBids {
			return nil, fmt.Errorf("order-id: [%s] agent bids sum [%d] not equal to order tokens [%d] ", orderId, agentBids, order.Tokens)
		}
		for producerId, tokens := range bids {
			capType := producers[producerId].CapacityType
			required, ok := order.Required[capType]
			if !ok {
				return nil, fmt.Errorf("order [%s] doesn't contain capacity type for producer [%s] with capacity type [%s]", orderId, producerId, capType)
			}
			result[producerId] = append(result[producerId], Bid{capType, required, tokens, orderId})
			logEvent("ordering.bid.created",
				slog.String("agentId", string(oa.id)),
				withOrderId(orderId),
				withProducerId(producerId),
				withCapacityType(capType),
				withCapacity(required),
				withTokens(tokens))
		}
	}
	oa.cmdHandled = true
	return result, nil
}

func NewOrderingAgent(id OrderingAgentId) *OrderingAgent {
	return &OrderingAgent{id, make(map[OrderId]OrderInfo), false}
}

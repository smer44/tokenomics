package domain

type OrderStatus byte

const (
	OrderStatusFunding = iota
	OrderStatusPlacing
	OrderStatusInProgress
	OrderStatusSuccess
	OrderStatusFailed
	OrderStatusPartiallyFailed
)

type partStatus byte

const (
	unknown partStatus = iota
	processing
	rejected
	completed
)

type Order struct {
	status            OrderStatus
	id                OrderId
	tokens            Tokens
	ps                ProcessSheet
	parts             map[CapacityType]partStatus
	consumerRequest   *ConsumerRequest
	investmentRequest *InvestmentRequest
}

type OrderIdGenerator interface {
	New() OrderId
}

func newInvestmentOrder(id OrderId, ps ProcessSheet, request InvestmentRequest) *Order {
	parts := make(map[CapacityType]partStatus, len(ps.Require))
	for p := range ps.Require {
		parts[p] = unknown
	}
	return &Order{OrderStatusFunding, id, 0, ps, parts, nil, &request}
}

func newCustomerOrder(id OrderId, t Tokens, ps ProcessSheet, request ConsumerRequest) *Order {
	if t <= 0 {
		panic("tokens must be greater than 0")
	}
	parts := make(map[CapacityType]partStatus, len(ps.Require))
	for p := range ps.Require {
		parts[p] = unknown
	}
	return &Order{OrderStatusPlacing, id, t, ps, parts, &request, nil}
}

type OrderInfo struct {
	Id           OrderId
	Tokens       Tokens
	ProcessSheet ProcessSheet
}

func (o *Order) Info() OrderInfo {
	return OrderInfo{o.id, o.tokens, o.ps}
}

func (o *Order) AgentId() OrderingAgentId {
	if o.investmentRequest != nil {
		return FromProducerId(o.investmentRequest.ProducerId)
	}
	return FromConsumerId(o.consumerRequest.ConsumerId)
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) CutOffPrice() CapacityUnitPrice {
	if o.investmentRequest == nil {
		panic("not an investement order")
	}
	return o.investmentRequest.CutOffPrice
}

func (o *Order) Fund(t Tokens) {
	if o.status != OrderStatusFunding {
		panic("not in funding status")
	}
	if o.investmentRequest == nil {
		panic("not an investement order")
	}
	o.status = OrderStatusPlacing
	o.tokens = t
}

func (o *Order) getPartStatus(bid Bid) partStatus {
	if bid.OrderId != o.id {
		panic("wrong order id")
	}
	status, ok := o.parts[bid.CapacityType]
	if !ok {
		panic(ErrNotFound)
	}
	return status
}

func (o *Order) Processing(bid Bid) {
	status := o.getPartStatus(bid)
	if status == processing {
		return
	}
	if status == completed {
		panic(ErrWrongState)
	}
	o.tokens -= bid.Tokens
	o.parts[bid.CapacityType] = processing
}

func (o *Order) Completed(bid Bid) {
	status := o.getPartStatus(bid)
	if status == completed {
		panic(ErrWrongState)
	}
	if status != processing {
		o.tokens -= bid.Tokens
	}
	o.parts[bid.CapacityType] = completed
}

func (o *Order) Rejected(bid Bid) {
	status := o.getPartStatus(bid)
	if status == processing || status == completed {
		panic(ErrWrongState)
	}
	o.parts[bid.CapacityType] = rejected
}

func (o *Order) EndTurn() (Score, bool) {

}

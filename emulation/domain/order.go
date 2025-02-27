package domain

import "errors"

type partStatus byte

const (
	unknown partStatus = iota
	processing
	rejected
	completed
)

type part struct {
	capacity Capacity
	status   partStatus
}

type Order struct {
	id                OrderId
	tokens            Tokens
	parts             map[CapacityType]*part
	consumerRequest   *ConsumerRequest
	investmentRequest *InvestmentRequest
	cycleCounter      uint
	funded            bool
}

type Score uint

type OrderIdGenerator interface {
	New() OrderId
}

func NewInvestmentOrder(id OrderId, ps ProcessSheet, request InvestmentRequest) *Order {
	parts := make(map[CapacityType]*part, len(ps.Require))
	for t, capacity := range ps.Require {
		parts[t] = &part{capacity, unknown}
	}
	return &Order{id, 0, parts, nil, &request, 0, false}
}

func NewConsumerOrder(id OrderId, ps ProcessSheet, request ConsumerRequest) *Order {
	parts := make(map[CapacityType]*part, len(ps.Require))
	for t, capacity := range ps.Require {
		parts[t] = &part{capacity, unknown}
	}
	return &Order{id, request.Tokens, parts, &request, nil, 0, true}
}

type OrderInfo struct {
	Id       OrderId
	Tokens   Tokens
	Required map[CapacityType]Capacity
}

var ErrFundingRequired = errors.New("funding required")

func (o *Order) mustBeFunded() {
	if o.RequiresFunding() {
		panic(ErrFundingRequired)
	}
}

func (o *Order) Info() OrderInfo {
	o.mustBeFunded()
	required := make(map[CapacityType]Capacity, len(o.parts))
	for k, v := range o.parts {
		if v.status == processing || v.status == completed {
			continue
		}
		required[k] = v.capacity
	}
	return OrderInfo{o.id, o.tokens, required}
}

func (o *Order) AgentId() OrderingAgentId {
	o.mustBeFunded()
	if o.investmentRequest != nil {
		return FromProducerId(o.investmentRequest.ProducerId)
	}
	return FromConsumerId(o.consumerRequest.ConsumerId)
}

func (o *Order) RequiresFunding() bool {
	return !o.funded
}

func (o *Order) CutOffPrice() CapacityUnitPrice {
	if o.investmentRequest == nil {
		panic("not an investement order")
	}
	return o.investmentRequest.CutOffPrice
}

func (o *Order) Fund(t Tokens) {
	if !o.RequiresFunding() {
		panic("not in funding status")
	}
	if o.investmentRequest == nil {
		panic("not an investement order")
	}
	o.tokens = t
	o.funded = true
}

func (o *Order) getPartStatus(ct CapacityType) partStatus {
	part, ok := o.parts[ct]
	if !ok {
		panic(ErrNotFound)
	}
	return part.status
}

func (o *Order) spendTokens(t Tokens) {
	if o.tokens < t {
		panic("too few tokens left")
	}
	o.tokens -= t
}

func (o *Order) Processing(ct CapacityType, t Tokens) {
	o.mustBeFunded()
	status := o.getPartStatus(ct)
	if status == processing {
		return
	}
	if status == completed {
		panic(ErrWrongState)
	}
	o.spendTokens(t)
	o.parts[ct].status = processing
}

func (o *Order) Completed(ct CapacityType, t Tokens) {
	o.mustBeFunded()
	status := o.getPartStatus(ct)
	if status == completed {
		panic(ErrWrongState)
	}
	if status != processing {
		o.spendTokens(t)
	}
	o.parts[ct].status = completed
}

func (o *Order) Rejected(ct CapacityType) {
	o.mustBeFunded()
	status := o.getPartStatus(ct)
	if status == processing || status == completed {
		panic(ErrWrongState)
	}
	o.parts[ct].status = rejected
}

type OrderEvent any
type ConsumerRequestCompleted struct {
	Request *ConsumerRequest
}
type InvestmentRequestCompleted struct {
	Request *InvestmentRequest
}
type ConsumerRequestRejected struct {
	Remaining Tokens
	Request   *ConsumerRequest
}
type InvestmentRequestRejected struct {
	Request *InvestmentRequest
}
type OrderStillProcessing struct {
}

func (o *Order) EndCycle() (Score, OrderEvent) {
	o.mustBeFunded()
	rejectedCount := 0
	completedCount := 0
	for _, part := range o.parts {
		switch part.status {
		case rejected:
			rejectedCount++
		case completed:
			completedCount++
		}
	}
	// completed
	if completedCount == len(o.parts) {
		scores := Score(0)
		if o.consumerRequest != nil {
			return scores, ConsumerRequestCompleted{o.consumerRequest}
		}
		return scores, InvestmentRequestCompleted{o.investmentRequest}
	}
	if rejectedCount == len(o.parts) {
		scores := Score(3)
		if o.consumerRequest != nil {
			return scores, ConsumerRequestRejected{o.tokens, o.consumerRequest}
		}
		return scores, InvestmentRequestRejected{o.investmentRequest}
	}
	if o.cycleCounter == 2 {
		scores := Score(5)
		if o.consumerRequest != nil {
			return scores, ConsumerRequestRejected{o.tokens, o.consumerRequest}
		}
		return scores, InvestmentRequestRejected{o.investmentRequest}
	}
	o.cycleCounter++
	return Score(1), OrderStillProcessing{}
}

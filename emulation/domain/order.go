package domain

import (
	"errors"
	"log/slog"
)

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
	order := &Order{id, 0, parts, nil, &request, 0, false}
	logEvent("order.investment.created",
		withOrderId(id),
		withProducerId(request.ProducerId),
		withProduct(request.Product))
	return order
}

func NewConsumerOrder(id OrderId, ps ProcessSheet, request ConsumerRequest) *Order {
	parts := make(map[CapacityType]*part, len(ps.Require))
	for t, capacity := range ps.Require {
		parts[t] = &part{capacity, unknown}
	}
	order := &Order{id, request.Tokens, parts, &request, nil, 0, true}
	logEvent("order.consumer.created",
		withOrderId(id),
		withConsumerId(request.ConsumerId),
		withProduct(request.Product),
		withTokens(request.Tokens))
	return order
}

type OrderInfo struct {
	Id       OrderId
	Tokens   Tokens
	Required map[CapacityType]Capacity
}

func (i OrderInfo) Fulfilled() bool {
	return len(i.Required) == 0
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
	logEvent("order.funded",
		withOrderId(o.id),
		withTokens(t),
		withProducerId(o.investmentRequest.ProducerId))
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
	logEvent("order.tokens.spent",
		withOrderId(o.id),
		withTokens(t),
		slog.Int("remainingTokens", int(o.tokens)))
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
	logEvent("order.part.processing",
		withOrderId(o.id),
		withCapacityType(ct),
		withTokens(t))
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
	logEvent("order.part.completed",
		withOrderId(o.id),
		withCapacityType(ct),
		withTokens(t))
}

func (o *Order) Rejected(ct CapacityType) {
	o.mustBeFunded()
	status := o.getPartStatus(ct)
	if status == processing || status == completed {
		panic(ErrWrongState)
	}
	o.parts[ct].status = rejected
	logEvent("order.part.rejected",
		withOrderId(o.id),
		withCapacityType(ct))
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

func (o *Order) CompleteCycle() (Score, OrderEvent) {
	o.mustBeFunded()
	rejectedCount := 0
	completedCount := 0
	for ct, part := range o.parts {
		switch part.status {
		case rejected:
			rejectedCount++
			logEvent("order.cycle.part.rejected",
				withOrderId(o.id),
				withCapacityType(ct))
		case completed:
			completedCount++
			logEvent("order.cycle.part.completed",
				withOrderId(o.id),
				withCapacityType(ct))
		}
	}

	logEvent("order.cycle.status",
		withOrderId(o.id),
		slog.Int("completedParts", completedCount),
		slog.Int("rejectedParts", rejectedCount),
		slog.Int("totalParts", len(o.parts)),
		slog.Uint64("cycleCounter", uint64(o.cycleCounter)))

	// completed
	if completedCount == len(o.parts) {
		scores := Score(0)
		if o.consumerRequest != nil {
			logEvent("order.cycle.completed.consumer",
				withOrderId(o.id),
				withConsumerId(o.consumerRequest.ConsumerId),
				withProduct(o.consumerRequest.Product))
			return scores, ConsumerRequestCompleted{o.consumerRequest}
		}
		logEvent("order.cycle.completed.investment",
			withOrderId(o.id),
			withProducerId(o.investmentRequest.ProducerId),
			withProduct(o.investmentRequest.Product))
		return scores, InvestmentRequestCompleted{o.investmentRequest}
	}
	if rejectedCount == len(o.parts) {
		scores := Score(3)
		if o.consumerRequest != nil {
			logEvent("order.cycle.rejected.consumer",
				withOrderId(o.id),
				withConsumerId(o.consumerRequest.ConsumerId),
				withTokens(o.tokens))
			return scores, ConsumerRequestRejected{o.tokens, o.consumerRequest}
		}
		logEvent("order.cycle.rejected.investment",
			withOrderId(o.id),
			withProducerId(o.investmentRequest.ProducerId))
		return scores, InvestmentRequestRejected{o.investmentRequest}
	}
	if o.cycleCounter == 2 {
		scores := Score(5)
		if o.consumerRequest != nil {
			logEvent("order.cycle.timeout.consumer",
				withOrderId(o.id),
				withConsumerId(o.consumerRequest.ConsumerId),
				withTokens(o.tokens))
			return scores, ConsumerRequestRejected{o.tokens, o.consumerRequest}
		}
		logEvent("order.cycle.timeout.investment",
			withOrderId(o.id),
			withProducerId(o.investmentRequest.ProducerId))
		return scores, InvestmentRequestRejected{o.investmentRequest}
	}
	o.cycleCounter++
	logEvent("order.cycle.processing",
		withOrderId(o.id),
		slog.Uint64("cycleCounter", uint64(o.cycleCounter)))
	return Score(1), OrderStillProcessing{}
}

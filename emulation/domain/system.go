package domain

import (
	"errors"
	"log/slog"

	"github.com/samber/lo"
)

var ErrNotFound = errors.New("not found")
var ErrWrongState = errors.New("wrong state")

type SystemState byte

const (
	SystemStateOrdersPlacement SystemState = iota
	SystemStateOrdering
)

type OrderingAgentId string

type SystemInfo struct {
	State        SystemState
	CycleCounter uint
}

type OrderingAgentInfo struct {
	Id OrderingAgentId
}

func FromProducerId(id ProducerId) OrderingAgentId {
	return OrderingAgentId(string(id))
}
func FromConsumerId(id ConsumerId) OrderingAgentId {
	return OrderingAgentId(string(id))
}

type System struct {
	state           SystemState
	idGen           OrderIdGenerator
	cycleEmission   Tokens
	investmentFund  Tokens
	processSheets   map[Product]ProcessSheet
	producerLookup  map[CapacityType][]ProducerId
	producingAgents map[ProducerId]*ProducingAgent
	producerInfos   map[ProducerId]ProducerInfo
	orderingAgents  map[OrderingAgentId]*OrderingAgent
	orders          map[OrderId]*Order
	consumers       map[ConsumerId]Consumer
	cycleCounter    uint
}

func NewSystem(idGen OrderIdGenerator, config *Configuration, consumers map[ConsumerId]Consumer) *System {
	s := &System{
		SystemStateOrdersPlacement,
		idGen,
		config.CycleEmission,
		0,
		lo.SliceToMap(config.ProcessSheets, func(ps ProcessSheet) (Product, ProcessSheet) {
			return ps.Product, ps
		}),
		lo.SliceToMap(config.ProducerConfigs, func(p ProducingAgentConfig) (CapacityType, []ProducerId) {
			return p.Type, []ProducerId{p.Id}
		}),
		lo.SliceToMap(config.ProducerConfigs, func(p ProducingAgentConfig) (ProducerId, *ProducingAgent) {
			return p.Id, newProducingAgent(p)
		}),
		nil,
		map[OrderingAgentId]*OrderingAgent{},
		map[OrderId]*Order{},
		consumers,
		0,
	}
	s.producerInfos = lo.MapEntries(s.producingAgents, func(id ProducerId, ps *ProducingAgent) (ProducerId, ProducerInfo) {
		return id, ps.Info()
	})
	for _, c := range consumers {
		id := FromConsumerId(c.Id())
		s.orderingAgents[id] = NewOrderingAgent(id)
	}
	for prodId := range s.producingAgents {
		id := FromProducerId(prodId)
		s.orderingAgents[id] = NewOrderingAgent(id)
	}
	s.startCycle()
	return s
}

func (s *System) OrderingAgentView(id OrderingAgentId) (OrderingAgentView, error) {
	if s.state != SystemStateOrdering {
		return OrderingAgentView{}, ErrWrongState
	}
	agent, ok := s.orderingAgents[id]
	if !ok {
		return OrderingAgentView{}, ErrNotFound
	}
	return agent.View(s.producerInfos), nil
}

func (s *System) OrderingAgentAction(id OrderingAgentId, cmd OrderingAgentCommand) error {
	if s.state != SystemStateOrdering {
		return ErrWrongState
	}
	agent, ok := s.orderingAgents[id]
	if !ok {
		return ErrNotFound
	}
	bids, err := agent.HandleCmd(cmd, s.producerInfos)
	if err != nil {
		return err
	}
	for prodId, bids := range bids {
		p, ok := s.producingAgents[prodId]
		if !ok {
			return ErrNotFound
		}
		p.PlaceBids(bids)
	}
	return nil
}

func (s *System) ProducingAgentView(id ProducerId) (ProducingAgentView, error) {
	if s.state != SystemStateOrdersPlacement {
		return ProducingAgentView{}, ErrWrongState
	}
	p, ok := s.producingAgents[id]
	if !ok {
		return ProducingAgentView{}, ErrNotFound
	}
	return p.View(), nil
}

func (s *System) ProducingAgentAction(id ProducerId, cmd ProducingAgentCommand) error {
	if s.state != SystemStateOrdersPlacement {
		return ErrWrongState
	}
	p, ok := s.producingAgents[id]
	if !ok {
		return ErrNotFound
	}
	investmentRequests, err := p.HandleCmd(cmd)
	if err != nil {
		return err
	}
	for _, r := range investmentRequests {
		id := s.idGen.New()
		s.orders[id] = NewInvestmentOrder(id, MustGet(s.processSheets, r.Product), r)
	}
	return nil
}

func (s *System) placeComsumersOrders() {
	for _, c := range s.consumers {
		for _, request := range c.Order() {
			id := s.idGen.New()
			order := NewConsumerOrder(id, MustGet(s.processSheets, request.Product), request)
			s.orders[id] = order
			logEvent("system.order.placed",
				withOrderId(id),
				withConsumerId(request.ConsumerId),
				withProduct(request.Product),
				withTokens(request.Tokens))
		}
	}
}

func (s *System) emit() {
	s.investmentFund = s.cycleEmission / 2
	if len(s.consumers) == 0 {
		return
	}
	consumerTokens := s.cycleEmission / 2 / Tokens(len(s.consumers))
	logEvent("system.tokens.emitted",
		withTokens(s.cycleEmission),
		withTokens(s.investmentFund),
		withTokens(consumerTokens),
		slog.Int("consumers", len(s.consumers)))
	for _, c := range s.consumers {
		c.Emit(consumerTokens)
	}
}

func (s *System) startCycle() {
	s.state = SystemStateOrdersPlacement
	logEvent("system.cycle.started",
		withState(s.state),
		withCycleCounter(s.cycleCounter))
	s.emit()
	s.placeComsumersOrders()
	s.cycleCounter++
}

func distibuteInvestmentFund(orders map[OrderId]*Order, investmentFund Tokens) {
	// Make funding
	// Distibute investment fund accodingly producer's cut off price  (capacity deficit)
	totalCutOffSum := CapacityUnitPrice(0)
	nanCount := 0
	unfundedOrders := 0

	logEvent("system.investment.distribution.started",
		withTokens(investmentFund))

	for _, order := range orders {
		if !order.RequiresFunding() {
			continue
		}
		unfundedOrders++
		if !order.CutOffPrice().IsNaN() {
			totalCutOffSum += order.CutOffPrice()
		} else {
			nanCount++
		}
	}

	logEvent("system.investment.distribution.calculated",
		slog.Float64("totalCutOffSum", float64(totalCutOffSum)),
		slog.Int("nanPriceOrders", nanCount),
		slog.Int("unfundedOrders", unfundedOrders))

	remains := investmentFund
	for orderId, order := range orders {
		if !order.RequiresFunding() || order.CutOffPrice().IsNaN() {
			continue
		}
		funds := Tokens(float32(order.CutOffPrice()) * float32(investmentFund) / float32(totalCutOffSum))
		order.Fund(funds)
		remains -= funds

		logEvent("system.investment.order.funded",
			withOrderId(orderId),
			withTokens(funds),
			withCutOffPrice(order.CutOffPrice()))
	}
	if nanCount == 0 {
		if remains > 0 {
			logEvent("system.investment.distribution.completed",
				withTokens(remains),
				slog.String("status", "withRemainder"))
		} else {
			logEvent("system.investment.distribution.completed",
				slog.String("status", "fullyDistributed"))
		}
		return
	}

	funds := remains / Tokens(nanCount)
	for orderId, order := range orders {
		if !order.RequiresFunding() {
			continue
		}
		order.Fund(funds)
		remains -= funds

		logEvent("system.investment.order.funded.nan",
			withOrderId(orderId),
			withTokens(funds))
	}

	logEvent("system.investment.distribution.completed",
		withTokens(remains),
		slog.String("status", "nanDistributed"))
}

func (s *System) StartOrdering() error {
	if s.state != SystemStateOrdersPlacement {
		return ErrWrongState
	}

	logEvent("system.ordering.started",
		withState(s.state),
		withTokens(s.investmentFund),
		slog.Int("orders", len(s.orders)))

	distibuteInvestmentFund(s.orders, s.investmentFund)

	// Place orders
	ordersByAgent := make(map[OrderingAgentId]int)
	for _, order := range s.orders {
		agentId := order.AgentId()
		ordersByAgent[agentId]++
		MustGet(s.orderingAgents, agentId).PlaceOrder(order.Info())
	}

	for agentId, count := range ordersByAgent {
		logEvent("system.orders.distributed",
			slog.String("agentId", string(agentId)),
			slog.Int("orderCount", count))
	}

	s.state = SystemStateOrdering

	logEvent("system.ordering.state.changed",
		withState(s.state))

	return nil
}

type CycleResult struct {
	Score Score
}

func (s *System) CompleteCycle() (CycleResult, error) {
	if s.state != SystemStateOrdering {
		return CycleResult{}, ErrWrongState
	}

	logEvent("system.cycle.completing",
		withCycleCounter(s.cycleCounter))

	for _, oa := range s.orderingAgents {
		oa.CompleteCycle()
	}

	for _, p := range s.producingAgents {
		result := p.Produce()
		for _, bid := range result.Processing {
			MustGet(s.orders, bid.OrderId).Processing(bid.CapacityType, bid.Tokens)
			logEvent("system.order.processing",
				withOrderId(bid.OrderId),
				withCapacityType(bid.CapacityType),
				withTokens(bid.Tokens))
		}
		for _, bid := range result.Completed {
			MustGet(s.orders, bid.OrderId).Completed(bid.CapacityType, bid.Tokens)
			logEvent("system.order.completed",
				withOrderId(bid.OrderId),
				withCapacityType(bid.CapacityType),
				withTokens(bid.Tokens))
		}
		for _, bid := range result.Rejected {
			MustGet(s.orders, bid.OrderId).Rejected(bid.CapacityType)
			logEvent("system.order.rejected",
				withOrderId(bid.OrderId),
				withCapacityType(bid.CapacityType))
		}
	}

	cycleScore := Score(0)
	for id, order := range s.orders {
		score, event := order.CompleteCycle()
		cycleScore += score
		completed := true
		switch e := event.(type) {
		case ConsumerRequestCompleted:
			logEvent("system.request.completed.consumer",
				withOrderId(id),
				withConsumerId(e.Request.ConsumerId))
		case InvestmentRequestCompleted:
			logEvent("system.request.completed.investment",
				withOrderId(id),
				withProducerId(e.Request.ProducerId))
			s.producingAgents[e.Request.ProducerId].InvesetmentCompleted(e.Request)
		case ConsumerRequestRejected:
			logEvent("system.request.rejected.consumer",
				withOrderId(id),
				withConsumerId(e.Request.ConsumerId),
				withTokens(e.Remaining))
			s.consumers[e.Request.ConsumerId].Emit(e.Remaining)
		case InvestmentRequestRejected:
			logEvent("system.request.rejected.investment",
				withOrderId(id),
				withProducerId(e.Request.ProducerId))
			s.producingAgents[e.Request.ProducerId].InvesetmentRejected(e.Request)
		case OrderStillProcessing:
			completed = false
		default:
			panic(errors.ErrUnsupported)
		}
		if completed {
			delete(s.orders, id)
		}
	}

	logEvent("system.cycle.completed",
		withCycleCounter(s.cycleCounter),
		slog.Int("score", int(cycleScore)))

	s.startCycle()
	return CycleResult{cycleScore}, nil
}

func (s *System) GetProducerInfos() map[ProducerId]ProducerInfo {
	return s.producerInfos
}

func (s *System) GetOrderingAgentInfos() map[OrderingAgentId]OrderingAgentInfo {
	result := make(map[OrderingAgentId]OrderingAgentInfo, len(s.orderingAgents))
	for id := range s.orderingAgents {
		result[id] = OrderingAgentInfo{id}
	}
	return result
}

func (s *System) GetSystemInfo() SystemInfo {
	return SystemInfo{
		State:        s.state,
		CycleCounter: s.cycleCounter,
	}
}

package application

import (
	"emulation/domain"
	"emulation/models"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"
)

type sequentialGenerator uint64

func (s *sequentialGenerator) New() domain.OrderId {
	id := domain.OrderId(strconv.Itoa(int(*s)))
	(*s)++
	return id
}

type Emulator struct {
	idGen  sequentialGenerator
	rwMu   *sync.RWMutex
	system *domain.System
	config *domain.Configuration
}

func NewEmulator() *Emulator {
	slog.Info("emulator.initializing")
	e := &Emulator{0, &sync.RWMutex{}, nil, nil}
	e.Reset()
	return e
}

func (e *Emulator) Reset() {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

	slog.Info("emulator.reset.started")

	// Load configuration from JSON file
	configData, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config domain.Configuration
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	e.system = domain.NewSystem(&e.idGen, &config, map[domain.ConsumerId]domain.Consumer{})
	e.config = &config

	slog.Info("emulator.reset.completed",
		slog.Int("cycleEmission", int(config.CycleEmission)),
		slog.Int("processSheets", len(config.ProcessSheets)),
		slog.Int("producers", len(config.ProducerConfigs)))
}

func (e *Emulator) GetOrderingAgentView(id domain.OrderingAgentId) (domain.OrderingAgentView, error) {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()

	view, err := e.system.OrderingAgentView(id)
	if err != nil {
		slog.Error("emulator.get_ordering_agent_view.failed",
			slog.String("agentId", string(id)),
			slog.String("error", err.Error()))
		return view, err
	}

	slog.Info("emulator.get_ordering_agent_view.success",
		slog.String("agentId", string(id)),
		slog.Int("incomingOrders", len(view.Incoming)))
	return view, nil
}

func (e *Emulator) GetProducingAgentView(id domain.ProducerId) (domain.ProducingAgentView, error) {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()

	view, err := e.system.ProducingAgentView(id)
	if err != nil {
		slog.Error("emulator.get_producing_agent_view.failed",
			slog.String("producerId", string(id)),
			slog.String("error", err.Error()))
		return view, err
	}

	slog.Info("emulator.get_producing_agent_view.success",
		slog.String("producerId", string(id)),
		slog.Int("capacity", int(view.Capacity)),
		slog.Int("maxCapacity", int(view.MaxCapacity)))
	return view, nil
}

func (e *Emulator) OrderingAgentAction(id domain.OrderingAgentId, cmd domain.OrderingAgentCommand) error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

	slog.Info("emulator.ordering_agent_action.started",
		slog.String("agentId", string(id)),
		slog.Int("orders", len(cmd.Orders)))

	if err := e.system.OrderingAgentAction(id, cmd); err != nil {
		slog.Error("emulator.ordering_agent_action.failed",
			slog.String("agentId", string(id)),
			slog.String("error", err.Error()))
		return err
	}

	slog.Info("emulator.ordering_agent_action.completed",
		slog.String("agentId", string(id)))
	return nil
}

func (e *Emulator) ProducingAgentAction(id domain.ProducerId, cmd domain.ProducingAgentCommand) error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

	slog.Info("emulator.producing_agent_action.started",
		slog.String("producerId", string(id)),
		slog.Bool("doUpgrade", cmd.DoUpgrade),
		slog.Bool("doRestoration", cmd.DoRestoration))

	if err := e.system.ProducingAgentAction(id, cmd); err != nil {
		slog.Error("emulator.producing_agent_action.failed",
			slog.String("producerId", string(id)),
			slog.String("error", err.Error()))
		return err
	}

	slog.Info("emulator.producing_agent_action.completed",
		slog.String("producerId", string(id)))
	return nil
}

func (e *Emulator) StartOrdering() error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

	slog.Info("emulator.start_ordering.started")

	if err := e.system.StartOrdering(); err != nil {
		slog.Error("emulator.start_ordering.failed",
			slog.String("error", err.Error()))
		return err
	}

	slog.Info("emulator.start_ordering.completed")
	return nil
}

func (e *Emulator) CompleteCycle() (domain.CycleResult, error) {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

	slog.Info("emulator.complete_cycle.started")

	result, err := e.system.CompleteCycle()
	if err != nil {
		slog.Error("emulator.complete_cycle.failed",
			slog.String("error", err.Error()))
		return result, err
	}

	slog.Info("emulator.complete_cycle.completed",
		slog.Int("score", int(result.Score)))
	return result, nil
}

func (e *Emulator) GetProducerInfos() map[domain.ProducerId]domain.ProducerInfo {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()
	return e.system.GetProducerInfos()
}

func (e *Emulator) GetOrderingAgentInfos() map[domain.OrderingAgentId]domain.OrderingAgentInfo {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()
	return e.system.GetOrderingAgentInfos()
}

func (e *Emulator) GetSystemInfo() models.SystemInfo {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()
	info := e.system.GetSystemInfo()
	state := "OrdersPlacement"
	if info.State == domain.SystemStateOrdering {
		state = "Ordering"
	}
	return models.SystemInfo{
		State:        state,
		CycleCounter: int64(info.CycleCounter),
	}
}

func (e *Emulator) GetConfig() *domain.Configuration {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()
	return e.config
}

func (e *Emulator) UpdateConfig(config *domain.Configuration) error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

	slog.Info("emulator.update_config.started",
		slog.Int("cycleEmission", int(config.CycleEmission)),
		slog.Int("processSheets", len(config.ProcessSheets)),
		slog.Int("producers", len(config.ProducerConfigs)))

	e.config = config

	slog.Info("emulator.update_config.completed")
	return nil
}

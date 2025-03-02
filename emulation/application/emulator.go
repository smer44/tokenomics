package application

import (
	"emulation/domain"
	"emulation/models"
	"encoding/json"
	"log"
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
	e := &Emulator{0, &sync.RWMutex{}, nil, nil}
	e.Reset()
	return e
}

func (e *Emulator) Reset() {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()

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
}

func (e *Emulator) GetOrderingAgentView(id domain.OrderingAgentId) (domain.OrderingAgentView, error) {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()
	return e.system.OrderingAgentView(id)
}

func (e *Emulator) GetProducingAgentView(id domain.ProducerId) (domain.ProducingAgentView, error) {
	e.rwMu.RLock()
	defer e.rwMu.RUnlock()
	return e.system.ProducingAgentView(id)
}

func (e *Emulator) OrderingAgentAction(id domain.OrderingAgentId, cmd domain.OrderingAgentCommand) error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()
	return e.system.OrderingAgentAction(id, cmd)
}

func (e *Emulator) ProducingAgentAction(id domain.ProducerId, cmd domain.ProducingAgentCommand) error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()
	return e.system.ProducingAgentAction(id, cmd)
}

func (e *Emulator) StartOrdering() error {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()
	return e.system.StartOrdering()
}

func (e *Emulator) CompleteCycle() (domain.CycleResult, error) {
	e.rwMu.Lock()
	defer e.rwMu.Unlock()
	return e.system.CompleteCycle()
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
	e.config = config
	return nil
}

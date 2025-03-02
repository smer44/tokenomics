package domain

import (
	"fmt"
)

// ProcessSheet represents a production process that converts capacity into products
type ProcessSheet struct {
	Product Product                   `json:"product"`
	Require map[CapacityType]Capacity `json:"require"`
}

// Configuration represents the system configuration
type Configuration struct {
	CycleEmission   Tokens                 `json:"cycleEmission"`
	ProcessSheets   []ProcessSheet         `json:"processSheets"`
	ProducerConfigs []ProducingAgentConfig `json:"producerConfigs"`
}

func (c *Configuration) Validate() error {
	if c.CycleEmission <= 0 {
		return fmt.Errorf("cycle emission must be positive, got %d", c.CycleEmission)
	}

	// Validate process sheets
	processProducts := make(map[Product]bool)
	processCapacities := make(map[CapacityType]bool)
	for _, sheet := range c.ProcessSheets {
		if len(sheet.Require) == 0 {
			return fmt.Errorf("process sheet for product %v has no capacity requirements", sheet.Product)
		}
		if processProducts[sheet.Product] {
			return fmt.Errorf("duplicate product %v in process sheets", sheet.Product)
		}
		processProducts[sheet.Product] = true

		for capType, cap := range sheet.Require {
			if cap <= 0 {
				return fmt.Errorf("capacity requirement must be positive, got %d for type %s", cap, capType)
			}
			processCapacities[capType] = true
		}
	}

	// Validate producer configs
	producerIds := make(map[ProducerId]bool)
	producerCapTypes := make(map[CapacityType]bool)
	for _, config := range c.ProducerConfigs {
		if producerIds[config.Id] {
			return fmt.Errorf("duplicate producer id %s", config.Id)
		}
		producerIds[config.Id] = true

		if config.Capacity <= 0 {
			return fmt.Errorf("producer %s capacity must be positive, got %d", config.Id, config.Capacity)
		}

		producerCapTypes[config.Type] = true
	}

	// Cross-validate process sheets and producers
	for capType := range processCapacities {
		if !producerCapTypes[capType] {
			return fmt.Errorf("capacity type %s required by process sheets but no producer provides it", capType)
		}
	}

	return nil
}

// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProducingAgentView producing agent view
//
// swagger:model ProducingAgentView
type ProducingAgentView struct {

	// Current capacity. Degradates each turn. Could be increased to MaxCapacity with Restoration purchase
	Capacity int64 `json:"capacity,omitempty"`

	// Capacity decrease in the current cycle
	Degradation int64 `json:"degradation,omitempty"`

	// Agent ID
	ID string `json:"id,omitempty"`

	// Maximum capacity. Can be increased with Upgrade purchase
	MaxCapacity int64 `json:"maxCapacity,omitempty"`

	// Total capacity was requested in the previous cycle
	RequestedCapacity int64 `json:"requestedCapacity,omitempty"`

	// Capacity gain with Restoration
	Restoration int64 `json:"restoration,omitempty"`

	// Indicates Restoration production is running
	RestorationRunning bool `json:"restorationRunning,omitempty"`

	// MaxCapacity and Capacity gain with Upgrade
	Upgrade int64 `json:"upgrade,omitempty"`

	// Indicates Upgrade production is running
	UpgradeRunning bool `json:"upgradeRunning,omitempty"`
}

// Validate validates this producing agent view
func (m *ProducingAgentView) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this producing agent view based on context it is used
func (m *ProducingAgentView) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProducingAgentView) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProducingAgentView) UnmarshalBinary(b []byte) error {
	var res ProducingAgentView
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

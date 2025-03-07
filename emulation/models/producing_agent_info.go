// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProducingAgentInfo producing agent info
//
// swagger:model ProducingAgentInfo
type ProducingAgentInfo struct {

	// Current capacity value
	Capacity int64 `json:"capacity,omitempty"`

	// Capacity type
	CapacityType string `json:"capacityType,omitempty"`

	// The cut off price in the previous cycle
	CutOffPrice int64 `json:"cutOffPrice,omitempty"`

	// Agent ID
	ID string `json:"id,omitempty"`

	// Maximum capacity value
	MaxCapacity int64 `json:"maxCapacity,omitempty"`
}

// Validate validates this producing agent info
func (m *ProducingAgentInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this producing agent info based on context it is used
func (m *ProducingAgentInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProducingAgentInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProducingAgentInfo) UnmarshalBinary(b []byte) error {
	var res ProducingAgentInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

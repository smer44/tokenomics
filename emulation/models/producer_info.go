// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProducerInfo producer info
//
// swagger:model ProducerInfo
type ProducerInfo struct {

	// Current capacity value
	Capacity int64 `json:"Capacity,omitempty"`

	// Capacity type
	CapacityType int64 `json:"CapacityType,omitempty"`

	// The cut off price in the previous cycle
	CutOffPrice int64 `json:"CutOffPrice,omitempty"`

	// Producer agent ID
	ID string `json:"Id,omitempty"`

	// Maximum capacity value
	MaxCapacity int64 `json:"MaxCapacity,omitempty"`
}

// Validate validates this producer info
func (m *ProducerInfo) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this producer info based on context it is used
func (m *ProducerInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProducerInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProducerInfo) UnmarshalBinary(b []byte) error {
	var res ProducerInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

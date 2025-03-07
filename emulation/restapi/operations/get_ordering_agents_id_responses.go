// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"emulation/models"
)

// GetOrderingAgentsIDOKCode is the HTTP code returned for type GetOrderingAgentsIDOK
const GetOrderingAgentsIDOKCode int = 200

/*
GetOrderingAgentsIDOK Successful response

swagger:response getOrderingAgentsIdOK
*/
type GetOrderingAgentsIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.OrderingAgentView `json:"body,omitempty"`
}

// NewGetOrderingAgentsIDOK creates GetOrderingAgentsIDOK with default headers values
func NewGetOrderingAgentsIDOK() *GetOrderingAgentsIDOK {

	return &GetOrderingAgentsIDOK{}
}

// WithPayload adds the payload to the get ordering agents Id o k response
func (o *GetOrderingAgentsIDOK) WithPayload(payload *models.OrderingAgentView) *GetOrderingAgentsIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get ordering agents Id o k response
func (o *GetOrderingAgentsIDOK) SetPayload(payload *models.OrderingAgentView) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetOrderingAgentsIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

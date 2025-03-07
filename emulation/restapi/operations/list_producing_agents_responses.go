// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"emulation/models"
)

// ListProducingAgentsOKCode is the HTTP code returned for type ListProducingAgentsOK
const ListProducingAgentsOKCode int = 200

/*
ListProducingAgentsOK OK

swagger:response listProducingAgentsOK
*/
type ListProducingAgentsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.ProducingAgentInfo `json:"body,omitempty"`
}

// NewListProducingAgentsOK creates ListProducingAgentsOK with default headers values
func NewListProducingAgentsOK() *ListProducingAgentsOK {

	return &ListProducingAgentsOK{}
}

// WithPayload adds the payload to the list producing agents o k response
func (o *ListProducingAgentsOK) WithPayload(payload []*models.ProducingAgentInfo) *ListProducingAgentsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list producing agents o k response
func (o *ListProducingAgentsOK) SetPayload(payload []*models.ProducingAgentInfo) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListProducingAgentsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.ProducingAgentInfo, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

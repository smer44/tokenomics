// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"emulation/models"
)

// ListProducersOKCode is the HTTP code returned for type ListProducersOK
const ListProducersOKCode int = 200

/*
ListProducersOK OK

swagger:response listProducersOK
*/
type ListProducersOK struct {

	/*
	  In: Body
	*/
	Payload []*models.ProducerInfo `json:"body,omitempty"`
}

// NewListProducersOK creates ListProducersOK with default headers values
func NewListProducersOK() *ListProducersOK {

	return &ListProducersOK{}
}

// WithPayload adds the payload to the list producers o k response
func (o *ListProducersOK) WithPayload(payload []*models.ProducerInfo) *ListProducersOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the list producers o k response
func (o *ListProducersOK) SetPayload(payload []*models.ProducerInfo) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ListProducersOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.ProducerInfo, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

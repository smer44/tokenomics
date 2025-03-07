// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"emulation/models"
)

// GetSystemInfoOKCode is the HTTP code returned for type GetSystemInfoOK
const GetSystemInfoOKCode int = 200

/*
GetSystemInfoOK OK

swagger:response getSystemInfoOK
*/
type GetSystemInfoOK struct {

	/*
	  In: Body
	*/
	Payload []*models.SystemInfo `json:"body,omitempty"`
}

// NewGetSystemInfoOK creates GetSystemInfoOK with default headers values
func NewGetSystemInfoOK() *GetSystemInfoOK {

	return &GetSystemInfoOK{}
}

// WithPayload adds the payload to the get system info o k response
func (o *GetSystemInfoOK) WithPayload(payload []*models.SystemInfo) *GetSystemInfoOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get system info o k response
func (o *GetSystemInfoOK) SetPayload(payload []*models.SystemInfo) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSystemInfoOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.SystemInfo, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

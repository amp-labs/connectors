package interpreter

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var ErrCannotParseErrorResponse = errors.New("implementation cannot process error response")

// FaultyResponder is an implementation of FaultyResponseHandler.
// It uses common techniques to handle error response returned by provider.
type FaultyResponder struct {
	StatusCodeMapper

	errorSwitch *FormatSwitch
}

// NewFaultyResponder creates error responder that will be used when provider responds with erroneous payload.
//   - FormatSwitch will be used to select the best matching error format.
//     It can be null if you don't want pretty parsing and formating.
//   - StatusCodeMap will be used to enhance error message.
//     This is an optional map that will precede any default status to error mapping.
func NewFaultyResponder(errorSwitch *FormatSwitch, statusCodeMap map[int]error) *FaultyResponder {
	return &FaultyResponder{
		errorSwitch:      errorSwitch,
		StatusCodeMapper: StatusCodeMapper{Registry: statusCodeMap},
	}
}

func (r FaultyResponder) HandleErrorResponse(res *http.Response, body []byte) error {
	if r.errorSwitch == nil {
		return ErrCannotParseErrorResponse
	}

	// Locate best schema to describe response.
	schema := r.errorSwitch.ParseJSON(body)

	// Match status code to error. Enhance it with schema message.
	return schema.CombineErr(r.MatchStatusCodeError(res, body))
}

type StatusCodeMapper struct {
	Registry map[int]error
}

func (m StatusCodeMapper) MatchStatusCodeError(res *http.Response, body []byte) error {
	if m.Registry == nil {
		return DefaultStatusCodeMappingToErr(res, body)
	}

	// Check if status code was overridden.
	mappedErr, ok := m.Registry[res.StatusCode]
	if !ok {
		return DefaultStatusCodeMappingToErr(res, body)
	}

	return mappedErr
}

// DirectFaultyResponder is an implementation of FaultyResponseHandler.
// It is a simple wrapper that allows you to directly provide an implementation of handler callback.
// First consider FaultyResponder that reduces the boilerplate.
type DirectFaultyResponder struct {
	Callback func(res *http.Response, body []byte) error
}

func (r DirectFaultyResponder) HandleErrorResponse(res *http.Response, body []byte) error {
	if r.Callback == nil {
		return ErrCannotParseErrorResponse
	}

	return r.Callback(res, body)
}

// XMLFaultyResponder is an implementation of FaultyResponseHandler.
type XMLFaultyResponder struct {
	StatusCodeMapper

	templates Templates
}

// NewXMLFaultyResponder creates error responder that will be used when provider responds with erroneous payload.
//   - Templates will attempt to release XML data into each template until the first successful format.
//   - StatusCodeMap will be used to enhance error message.
//     This is an optional map that will precede any default status to error mapping.
func NewXMLFaultyResponder(templates Templates, statusCodeMap map[int]error) *XMLFaultyResponder {
	return &XMLFaultyResponder{
		templates:        templates,
		StatusCodeMapper: StatusCodeMapper{Registry: statusCodeMap},
	}
}

func (r XMLFaultyResponder) HandleErrorResponse(res *http.Response, body []byte) error {
	if len(r.templates) == 0 {
		return ErrCannotParseErrorResponse
	}

	// Choose the first XML template that can be unmarshalled.
	for _, template := range r.templates {
		if template == nil {
			// Nil templates are not allowed.
			return ErrCannotParseErrorResponse
		}

		tmpl := template()
		if err := xml.Unmarshal(body, &tmpl); err == nil {
			return fmt.Errorf("provider error: %w", tmpl.CombineErr(r.MatchStatusCodeError(res, body)))
		}
	}

	// Response body cannot be understood in the form of valid XML structure.
	// Default error handling.
	return common.InterpretError(res, body)
}

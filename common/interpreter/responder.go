package interpreter

import (
	"net/http"
)

// FaultyResponder is an implementation of FaultyResponseHandler.
// It uses common techniques to handle error response returned by provider.
type FaultyResponder struct {
	errorSwitch   *FormatSwitch
	statusCodeMap map[int]error
}

// NewFaultyResponder creates error responder that will be used when provider responds with erroneous payload.
//   - FormatSwitch will be used to select the best matching error format.
//     It can be null if you don't want pretty parsing and formating.
//   - StatusCodeMap will be used to enhance error message.
//     This is an optional map that will precede any default status to error mapping.
func NewFaultyResponder(errorSwitch *FormatSwitch, statusCodeMap map[int]error) *FaultyResponder {
	return &FaultyResponder{
		errorSwitch:   errorSwitch,
		statusCodeMap: statusCodeMap,
	}
}

func (r FaultyResponder) HandleErrorResponse(res *http.Response, body []byte) error {
	// Locate best schema to describe response.
	schema := r.errorSwitch.ParseJSON(body)

	// Match status code to error. Enhance it with schema message.
	return schema.CombineErr(r.matchStatusCodeError(res, body))
}

func (r FaultyResponder) matchStatusCodeError(res *http.Response, body []byte) error {
	if r.statusCodeMap == nil {
		return DefaultStatusCodeMappingToErr(res, body)
	}

	// Check if status code was overridden.
	mappedErr, ok := r.statusCodeMap[res.StatusCode]
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
	return r.Callback(res, body)
}

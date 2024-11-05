// nolint:ireturn
package interpreter

import (
	"encoding/json"
	"errors"
	"log/slog"
)

var ErrUnknownResponseFormat = errors.New("unknown response format")

// FormatSwitch allows to select the most appropriate format.
// Switch will traverse every template stopping at the closest match, which best describes server response.
// Then ErrorDescriptor will convert itself into a composite go error.
type FormatSwitch struct {
	// List of templates to choose from when parsing data.
	templates []FormatTemplate
}

func NewFormatSwitch(templates ...FormatTemplate) *FormatSwitch {
	return &FormatSwitch{
		templates: templates,
	}
}

// ParseJSON selects a template, populates it, and returns the result. If the error response is concise and clear,
// it’s used directly; otherwise, the entire response is used to build an error.
// This strategy ensures complete visibility into the provider’s response.
//
// The JSON data may either contain a single object, interpreted as an ErrorDescriptor,
// or an array of objects, with each mapped to matching ErrorDescriptor.
func (s FormatSwitch) ParseJSON(data []byte) ErrorDescriptor {
	// First: assume HTTP response data is an array of objects.
	list := make([]map[string]any, 0)
	if err := json.Unmarshal(data, &list); err == nil {
		return s.mapObjectListToErrorDescriptor(list)
	}

	// Second: assume HTTP response data is a single object.
	object := make(map[string]any)
	if err := json.Unmarshal(data, &object); err == nil {
		return s.mapObjectToErrorDescriptor(data, object)
	}

	// The response was likely not valid JSON format, neither object nor an array.
	// Returning a default raw descriptor.
	return defaultErrorDescriptor{
		responseData: data,
	}
}

// Response object will be matched against error format to produce error descriptor.
//
// Ex: provider response is object:
//
//	{
//		"code":128,
//		"message":"conflicting entries"
//	},
func (s FormatSwitch) mapObjectToErrorDescriptor(
	data []byte, jsonResponse map[string]any,
) ErrorDescriptor {
	for i := range s.templates {
		// explicit assignment because later we use pointer, this way it is not a pointer to a loop variable
		template := s.templates[i]

		if template.matches(jsonResponse) {
			// We found the perfect match.
			tmpl := template.Template()
			if err := json.Unmarshal(data, &tmpl); err == nil {
				// Successful parse.
				return tmpl
			}

			// Matched but couldn't parse. Did the server format change?
			// We will continue searching for the closest template as fallback.
			slog.Info("provider error response format has changed")
		}
	}

	// None of the templates describe the format.
	// Default fallback.
	return defaultErrorDescriptor{
		responseData: data,
	}
}

// We are dealing with the list of error objects.
// Each error object will be mapped to the template.
//
// Ex: provider response is array:
// [
//
//	{"code":128, "message":"conflicting entries"},
//	{"code":70, "message":"invalid time format", "description":"use ISO nano"},
//
// ].
func (s FormatSwitch) mapObjectListToErrorDescriptor(responses []map[string]any) ErrorDescriptor {
	descr := &listErrorDescriptor{
		list: make([]ErrorDescriptor, 0),
	}

	for _, rsp := range responses {
		data, err := json.Marshal(rsp)
		if err != nil {
			// this shouldn't happen, if we cannot marshal back we skip this json object.
			continue
		}

		descriptor := s.mapObjectToErrorDescriptor(data, rsp)
		descr.addErr(descriptor)
	}

	return descr
}

// Describes a list of errors, each using the same 'base' reason for consistency.
//
// For example, if the base is "bad request", errors like "invalid field" and "conflicting values"
// will both be associated with "bad request". This is not a linked list of nested errors, as "invalid field"
// did not cause "conflicting values", but both occur due to the same "bad request".
type listErrorDescriptor struct {
	list []ErrorDescriptor
}

func (d *listErrorDescriptor) addErr(descriptor ErrorDescriptor) {
	d.list = append(d.list, descriptor)
}

func (d *listErrorDescriptor) CombineErr(base error) error {
	list := make([]error, len(d.list))

	for i, descriptor := range d.list {
		list[i] = descriptor.CombineErr(base)
	}

	return errors.Join(list...)
}

type defaultErrorDescriptor struct {
	responseData []byte
}

func (d defaultErrorDescriptor) CombineErr(base error) error {
	return errors.Join(
		base,
		ErrUnknownResponseFormat,
		errors.New(string(d.responseData)), // nolint:goerr113
	)
}

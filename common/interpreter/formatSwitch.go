package interpreter

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrInvalidFormatExpectation = errors.New("used format template is not of valid type")
	ErrUnknownResponseFormat    = errors.New("unknown response format")
)

// FormatSwitch allows to select the most appropriate format.
// Switch will traverse every template stopping at the closest match, which best describes server response.
// Then common.ErrorDescriptor will convert itself into a composite go error.
type FormatSwitch struct {
	// List of templates to choose from when parsing data.
	templates []FormatTemplate
}

func NewFormatSwitch(templates ...FormatTemplate) *FormatSwitch {
	return &FormatSwitch{
		templates: templates,
	}
}

// ParseJSON will select one of the templates, populate and return it.
// If error response can be interpreted concisely and to the point, then great.
// Otherwise, fallback and use whole response to build an error.
// This strategy always exposes, never hides what the server sent us.
func (s FormatSwitch) ParseJSON(data []byte) (common.ErrorDescriptor, error) { // nolint:ireturn
	payload := make(map[string]any)
	if err := json.Unmarshal(data, &payload); err != nil {
		// The response was likely not valid JSON format.
		// Handling this error by returning default error description.
		return defaultErrorDescriptor{ // nolint:nilerr
			responseData: data,
		}, nil
	}

	for i := range s.templates {
		// explicit assignment because later we use pointer, this way it is not a pointer to a loop variable
		template := s.templates[i]

		if template.matches(payload) {
			// We found the perfect match.
			if err := json.Unmarshal(data, &template.Template); err == nil {
				// Successful parse.
				descriptor, ok := (template.Template).(common.ErrorDescriptor)
				if !ok {
					// Template doesn't satisfy output type expectation of FormatSwitch.
					// Connector implementation needs to fix this.
					// To resolve this ensure every Template is of common.ErrorDescriptor type.
					return nil, ErrInvalidFormatExpectation
				}

				return descriptor, nil
			}

			// Matched but couldn't parse. Did the server format change?
			// We will continue searching for the closest template as fallback.
			slog.Info("server error response format has changed")
		}
	}

	// None of the templates describe the format.
	// Default fallback.
	return defaultErrorDescriptor{
		responseData: data,
	}, nil
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
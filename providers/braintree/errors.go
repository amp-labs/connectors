package braintree

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// GraphQLError represents an error in the GraphQL response.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []string               `json:"path,omitempty"`
	Extensions map[string]any         `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents the location of an error in a GraphQL query.
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ResponseError represents the error structure in GraphQL responses.
type ResponseError struct {
	Errors []GraphQLError `json:"errors,omitempty"`
}

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		messages[i] = obj.Message
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}

// checkErrorInResponse checks if there are any errors in the GraphQL response.
func checkErrorInResponse(resp *ResponseError) error {
	if resp == nil || len(resp.Errors) == 0 {
		return nil
	}

	var errorMsg strings.Builder

	// Braintree GraphQL returns errors in a standard format
	// Map error classes to common error types
	for _, err := range resp.Errors {
		if err.Extensions != nil {
			if errorClass, ok := err.Extensions["errorClass"].(string); ok {
				switch errorClass {
				case "VALIDATION":
					return fmt.Errorf("%w: %s", common.ErrBadRequest, err.Message)
				case "AUTHENTICATION":
					return fmt.Errorf("%w: %s", common.ErrAccessToken, err.Message)
				case "NOT_FOUND":
					return fmt.Errorf("%w: %s", common.ErrNotFound, err.Message)
				case "FORBIDDEN":
					return fmt.Errorf("%w: %s", common.ErrForbidden, err.Message)
				}
			}
		}

		// If no error class, collect error message
		errorMsg.WriteString(err.Message + "; ")
	}

	return errors.New(strings.TrimSuffix(errorMsg.String(), "; ")) // nolint:goerr113
}

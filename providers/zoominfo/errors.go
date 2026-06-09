package zoominfo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// ZoomInfo returns errors in the JSON:API "errors" array shape, e.g.
//
//	{"errors":[{"status":"400","title":"Bad Request","detail":"at least one valid input criterion is required"}]}
//
// A simpler {"message": "..."} shape is also tolerated as a fallback.
var errorFormats = interpreter.NewFormatSwitch( //nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"errors"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &SimpleError{} },
		},
	}...,
)

var statusCodeMapping = map[int]error{ //nolint:gochecknoglobals
	http.StatusNotAcceptable:       common.ErrBadRequest,
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}

// ResponseError models the JSON:API error envelope.
type ResponseError struct {
	Errors []errorDetail `json:"errors"`
}

type errorDetail struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func (r ResponseError) CombineErr(base error) error {
	messages := make([]string, 0, len(r.Errors))

	for _, detail := range r.Errors {
		switch {
		case detail.Detail != "":
			messages = append(messages, detail.Detail)
		case detail.Title != "":
			messages = append(messages, detail.Title)
		}
	}

	if len(messages) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, "; "))
}

// SimpleError models a flat {"message": "..."} error body.
type SimpleError struct {
	Message string `json:"message"`
}

func (r SimpleError) CombineErr(base error) error {
	if r.Message == "" {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Message)
}

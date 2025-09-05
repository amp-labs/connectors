package blackbaud

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common/interpreter"
)

// Implement error abstraction layers to streamline provider error handling.
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

// ResponseError represents an error response from the blankbaud API.
// In this connector, based on my analysis, there are two types of error responses:
//
// 1. Example:
//
//	{
//	   "type": "urn:blackbaud:general-translation-error",
//	   "title": "One or more validation errors occurred.",
//	   "status": 400,
//	   "detail": "Error processing Infinity response: <HTML format>",
//	   "trace_id": "c93436f067024cafbd34a3477fe0baec",
//	   "span_id": "fc0101290b2ab96e"
//	}
//
// 2. Example:
//
//	{
//	   "statusCode": 404,
//	   "message": "Resource not found",
//	   "status": 404,
//	   "title": "Resource not found"
//	}
//
// Define a struct that can handle both types of error responses shown above.
type ResponseError struct {
	StatusCode int    `json:"statusCode,omitempty"`
	Message    string `json:"message,omitempty"`
	Status     int    `json:"status"`
	Title      string `json:"title"`
	Type       string `json:"type,omitempty"`
	Detail     string `json:"detail,omitempty"`
	TraceId    string `json:"trace_id,omitempty"` // nolint:tagliatelle,revive
	SpanId     string `json:"span_id,omitempty"`  // nolint:tagliatelle,revive
}

func (r ResponseError) CombineErr(base error) error {
	message := r.Message

	if r.Detail != "" {
		document, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(r.Detail)))
		if err != nil {
			// ignore HTML that cannot be understood
			return base
		}

		message = document.Find("StatusMessage").Text()
	}

	return fmt.Errorf("%w: %v", base, message)
}

package shared

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var ErrorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"o:errorDetails"},
			Template: func() interpreter.ErrorDescriptor { return &NetSuiteErrorDetails{} },
		},
	}...,
)

var StatusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusInternalServerError: common.ErrServer,
	http.StatusBadGateway:          common.ErrServer,
	http.StatusServiceUnavailable:  common.ErrServer,
	http.StatusGatewayTimeout:      common.ErrServer,
}

// NetSuiteErrorDetails represents NetSuite's specific error response format
// Example: {"type": "...", "title": "Bad Request", "status": 400, "o:errorDetails": [...]}
// nolint:tagliatelle
type NetSuiteErrorDetails struct {
	Type         string              `json:"type"`
	Title        string              `json:"title"`
	Status       int                 `json:"status"`
	ErrorDetails []NetSuiteErrorItem `json:"o:errorDetails"`
}

// NetSuiteErrorItem represents individual error details within NetSuite's error response.
type NetSuiteErrorItem struct {
	Detail          string `json:"detail"`
	ErrorQueryParam string `json:"o:errorQueryParam"`
	ErrorCode       string `json:"o:errorCode"`
}

func (d NetSuiteErrorDetails) CombineErr(base error) error {
	details := make([]string, len(d.ErrorDetails))
	for i, item := range d.ErrorDetails {
		details[i] = item.Detail
	}

	message := strings.Join(details, "; ")
	if len(message) == 0 {
		message = d.Title
	}

	return fmt.Errorf("%w: %v", base, message)
}

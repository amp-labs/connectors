package dynamicscrm

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &CRMResponseError{} },
		},
	}...,
)

type CRMResponseError struct {
	Err CRMError `json:"error"`
}

type CRMError struct {
	*EnhancedCRMError

	Code    string `json:"code"`
	Message string `json:"message"`
}

type EnhancedCRMError struct {
	HelpLink     string `json:"@Microsoft.PowerApps.CDS.HelpLink"`           // nolint:tagliatelle
	TraceText    string `json:"@Microsoft.PowerApps.CDS.TraceText"`          // nolint:tagliatelle
	InnerMessage string `json:"@Microsoft.PowerApps.CDS.InnerError.Message"` // nolint:tagliatelle
}

func (r CRMResponseError) CombineErr(base error) error {
	if len(r.Err.Message) > 0 {
		return fmt.Errorf("%w: %s", base, r.Err.Message)
	}

	return base
}

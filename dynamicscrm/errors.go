package dynamicscrm

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	formats := interpreter.NewFormatSwitch(
		[]interpreter.FormatTemplate{
			{
				MustKeys: nil,
				Template: &CRMResponseError{},
			},
		}...,
	)

	schema := formats.ParseJSON(body)

	return schema.CombineErr(interpreter.DefaultStatusCodeMappingToErr(res, body))
}

type CRMResponseError struct {
	Err CRMError `json:"error"`
}

type CRMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	*EnhancedCRMError
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

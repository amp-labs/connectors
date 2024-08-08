package interpreter

import (
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var ErrMissingContentType = errors.New("mime.ParseMediaType failed")

// FaultyResponseHandler used to parse erroneous response.
type FaultyResponseHandler interface {
	HandleErrorResponse(res *http.Response, body []byte) error
}

// ErrorHandler invokes a function that matches response media type with parse error, ex: JSON<->JsonParserMethod
// otherwise defaults to general error interpretation.
type ErrorHandler struct {
	JSON FaultyResponseHandler
	XML  FaultyResponseHandler
}

func (h ErrorHandler) Handle(res *http.Response, body []byte) error {
	mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMissingContentType, err)
	}

	if h.JSON != nil && mediaType == "application/json" {
		return h.JSON.HandleErrorResponse(res, body)
	}

	if h.XML != nil && (mediaType == "text/xml" || mediaType == "application/xml") {
		return h.XML.HandleErrorResponse(res, body)
	}

	return common.InterpretError(res, body)
}

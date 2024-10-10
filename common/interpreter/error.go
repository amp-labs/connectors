package interpreter

import (
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
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
	HTML FaultyResponseHandler
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

	if h.HTML != nil && (mediaType == "text/html" || mediaType == "application/html") {
		return h.HTML.HandleErrorResponse(res, body)
	}

	return common.InterpretError(res, body)
}

func (h ErrorHandler) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "errorHandler",
		Constructor: handy.Returner(h),
	}
}

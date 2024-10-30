package interpreter

import (
	"mime"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// FaultyResponseHandler used to parse erroneous response.
type FaultyResponseHandler interface {
	HandleErrorResponse(res *http.Response, body []byte) error
}

// ErrorHandler invokes a function that matches response media type with parse error, ex: JSON<->JsonParserMethod
// otherwise defaults to general opaque error interpretation.
type ErrorHandler struct {
	JSON FaultyResponseHandler
	XML  FaultyResponseHandler
	HTML FaultyResponseHandler
}

func (h ErrorHandler) Handle(res *http.Response, body []byte) error {
	mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err == nil {
		if h.JSON != nil && mediaType == "application/json" {
			return h.JSON.HandleErrorResponse(res, body)
		}

		if h.XML != nil && (mediaType == "text/xml" || mediaType == "application/xml") {
			return h.XML.HandleErrorResponse(res, body)
		}

		if h.HTML != nil && (mediaType == "text/html" || mediaType == "application/html") {
			return h.HTML.HandleErrorResponse(res, body)
		}
	}

	// Default fallback, which treats body as opaque string to produce golang error.
	return common.InterpretError(res, body)
}

package interpreter

import (
	"errors"
	"mime"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var ErrUnparseableHTTPResponse = errors.New("unparseable HTTP response")

// FaultyResponseHandler used to parse erroneous response.
type FaultyResponseHandler interface {
	HandleErrorResponse(res *http.Response, body []byte) error
}

// ErrorHandler invokes a function that matches response media type with parse error, ex: JSON<->JsonParserMethod
// otherwise defaults to general opaque error interpretation.
// UnknownMedia is a special handler that is used in case Content-Type is unavailable.
type ErrorHandler struct {
	JSON         FaultyResponseHandler
	XML          FaultyResponseHandler
	HTML         FaultyResponseHandler
	UnknownMedia FaultyResponseHandler
}

func (h ErrorHandler) Handle(res *http.Response, body []byte) error { // nolint:cyclop
	mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		// Media type is unknown, therefore we cannot select appropriate response handler.
		// Therefore, using fallback.
		if h.UnknownMedia != nil {
			return h.UnknownMedia.HandleErrorResponse(res, body)
		}

		err = common.InterpretError(res, body)

		return errors.Join(ErrUnparseableHTTPResponse, err)
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

	// Default fallback, which treats body as opaque string to produce golang error.
	return common.InterpretError(res, body)
}

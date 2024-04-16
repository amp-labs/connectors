package interpreter

import (
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrUnmarshal       = errors.New("unmarshal failed")
	MissingContentType = errors.New("mime.ParseMediaType failed")
)

// FaultyResponseHandler used to parse erroneous response.
type FaultyResponseHandler func(res *http.Response, body []byte) error

// ErrorHandler invokes a function that matches response media type with parse error, ex: JSON<->JsonParserMethod
// otherwise defaults to general error interpretation.
type ErrorHandler struct {
	JSON FaultyResponseHandler
}

func (h ErrorHandler) Handle(res *http.Response, body []byte) error {
	mediaType, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("%w: %w", MissingContentType, err)
	}

	if h.JSON != nil && mediaType == "application/json" {
		return h.JSON(res, body)
	}

	return common.InterpretError(res, body)
}

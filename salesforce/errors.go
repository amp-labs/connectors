package salesforce

import (
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrNotArray         = errors.New("records is not an array")
	ErrNotObject        = errors.New("record isn't an object")
	ErrNoFields         = errors.New("no fields specified")
	ErrNotString        = errors.New("nextRecordsUrl isn't a string")
	ErrNotBool          = errors.New("done isn't a boolean")
	ErrNotNumeric       = errors.New("totalSize isn't numeric")
	ErrMissingSubdomain = errors.New("missing Salesforce workspace name")
	ErrMissingClient    = errors.New("JSON http client not set")
)

func (c *Connector) interpretError(res *http.Response, body []byte) error {
	// TODO: handle salesforce errors in a more robust way. For now, we just
	// handle the basic HTTP status codes and nothing else.
	return common.InterpretError(res, body)
}

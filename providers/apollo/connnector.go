package apollo

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
)

type Connector struct {
	deep.Clients
	deep.EmptyCloser
}

type parameters struct {
	paramsbuilder.Client
}

func NewConnector(opts ...Option) (*Connector, error) {
	return deep.Connector[Connector, parameters](providers.Apollo, interpreter.ErrorHandler{}).Build(opts)
}

type operation string

// getAPIURL builds the url we can write/read data from
// Depending on the operation(read or write), some objects will need different endpoints.
// That's the sole purpose of the variable ops.
func (c *Connector) getAPIURL(objectName string, ops operation) (*urlbuilder.URL, error) {
	relativePath := strings.Join([]string{restAPIPrefix, objectName}, "/")

	url, err := urlbuilder.New(c.BaseURL(), relativePath)
	if err != nil {
		return nil, err
	}

	// If the given object uses search endpoint for Reading,
	// checks for the  method and makes the call.
	// currently we do not support routing to Search method.
	//
	if usesSearching(objectName) && ops == readOp {
		switch {
		case in(objectName, postSearchObjects):
			return nil, common.ErrOperationNotSupportedForObject
		// Objects opportunities & users do not use the POST method
		// The POST search reading limits do  not apply to them.
		case in(objectName, getSearchObjects):
			url.AddPath(searchingPath)
		}
	}

	return url, nil
}

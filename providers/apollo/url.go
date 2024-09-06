package apollo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var (
	restAPIPrefix string = "v1"   //nolint:gochecknoglobals
	pageQuery     string = "page" //nolint:gochecknoglobals
	pageSize      string = "100"  //nolint:gochecknoglobals
)

func (c *Connector) getReadURL(objectName string) (*urlbuilder.URL, error) {
	url, err := c.getAPIURL(objectName)
	if err != nil {
		return nil, err
	}

	// If the given object uses search endpoint for Reading,
	// checks for the  method and makes the call.
	// currently we do not support routing to Search method.
	//
	if usesSearching(objectName) {
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

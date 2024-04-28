package outreach

import (
	"context"
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
)

// Header for the content-Type in JSONAPISpecification
var JSONAPIContentTypeHeader = common.Header{
	Key:   "Content-Type",
	Value: "application/vnd.api+json",
}

type writeMethod func(context.Context, string, any, ...common.Header) (*common.JSONHTTPResponse, error)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write writeMethod

	URL, err := url.JoinPath(c.BaseURL, config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		URL, err = url.JoinPath(URL, config.RecordId)
		if err != nil {
			return nil, err
		}

		write = c.Client.Patch
	} else {
		write = c.Client.Post
	}

	// Outreach requires everything to be wrapped in a "data" object.
	data := make(map[string]interface{})
	data["data"] = config.RecordData

	res, err := write(ctx, URL, data, JSONAPIContentTypeHeader)
	if err != nil {
		return nil, err
	}

	// parse the result
	fmt.Println(res)

	return nil, nil

}

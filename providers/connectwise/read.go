package connectwise

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "1000"

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	c.clientIdHeader().ApplyToRequest(req)

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))

	if conditions, ok := makeReadCondition(params); ok {
		url.WithQueryParam("conditions", conditions)
	}

	return url, nil
}

// All objects react to the LastUpdated query even if no such field exists in the object format.
func makeReadCondition(params common.ReadParams) (string, bool) {
	conditions := make([]string, 0)

	if !params.Since.IsZero() {
		// Example:
		// 	LastUpdated = [2016-08-20T18:04:26Z]
		condition := fmt.Sprintf("LastUpdated >= [%v]", datautils.Time.FormatRFC3339inUTC(params.Since))
		conditions = append(conditions, condition)
	}

	if !params.Until.IsZero() {
		condition := fmt.Sprintf("LastUpdated <= [%v]", datautils.Time.FormatRFC3339inUTC(params.Until))
		conditions = append(conditions, condition)
	}

	return strings.Join(conditions, " AND "), len(conditions) != 0
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(resp,
		// "" is used, because root level of JSON is right away an array.
		common.MakeRecordsFunc(""),
		nextRecordsURL(resp),
		readhelper.MakeMarshaledDataFuncWithId(
			recordTransformer(params.ObjectName),
			readhelper.IdFieldQuery{Field: "id"},
		),
		params.Fields,
	)
}

// recordTransformer returns a RecordTransformer that
//
//	(1) lifts customFields from the nested ConnectWise response into the top-level record map.
//	(2) lifts communication items of a contact as virtual fields.
func recordTransformer(objectName string) common.RecordTransformer {
	return func(node *ajson.Node) (map[string]any, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, err
		}

		if err = attachCustomFields(node, root); err != nil {
			return nil, err
		}

		if objectName == objectNameContacts {
			if err = attachCommunicationItems(node, root); err != nil {
				return nil, err
			}
		}

		return root, nil
	}
}

func nextRecordsURL(resp *common.JSONHTTPResponse) common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		return httpkit.HeaderLink(resp, "next"), nil
	}
}

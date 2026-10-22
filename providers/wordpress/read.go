package wordpress

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = "100"

func (c Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if params.NextPage != "" {
		return http.NewRequestWithContext(ctx, http.MethodGet, params.NextPage.String(), nil)
	}

	// For first page, construct the URL.
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if objectsWithPagination.Has(params.ObjectName) {
		url.WithQueryParam("per_page", readhelper.PageSizeWithDefaultStr(params, defaultPageSize))
		url.WithQueryParam("page", "1")
	}

	if spec, ok := objectsToIncrSpec[params.ObjectName]; ok {
		if !params.Since.IsZero() {
			url.WithQueryParam(spec.since, datautils.Time.FormatRFC3339inUTC(params.Since))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam(spec.since, datautils.Time.FormatRFC3339inUTC(params.Until))
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

// Query params used to control pagination are "page" and "per_page".
var objectsWithPagination = datautils.NewSetFromList([]string{ // nolint:gochecknoglobals
	"activity-log-events",
	"blocks",
	"comments",
	"feedback",
	"font-collections",
	"font-families",
	"guidelines",
	"icon-collections",
	"icons",
	"jetpack-forms",
	"jp_pay_order",
	"jp_pay_product",
	"knowledge",
	"media",
	"menu-items",
	"menus",
	"navigation",
	"pages",
	"patterns",
	"posts",
	"tags",
	"users",
	"abilities",
	"wp_guideline_type",
	"wp_knowledge_type",
	"wp_pattern_category",
})

type incrSpec struct {
	since string
	until string
}

var objectsToIncrSpec = map[string]incrSpec{ // nolint:gochecknoglobals
	"activity-log-events": {"modified_after", "modified_before"},
	"blocks":              {"modified_after", "modified_before"},
	"comments":            {"after", "before"},
	"feedback":            {"modified_after", "modified_before"},
	"guidelines":          {"modified_after", "modified_before"},
	"jetpack-forms":       {"modified_after", "modified_before"},
	"jp_pay_order":        {"modified_after", "modified_before"},
	"jp_pay_product":      {"modified_after", "modified_before"},
	"knowledge":           {"modified_after", "modified_before"},
	"media":               {"modified_after", "modified_before"},
	"menu-items":          {"modified_after", "modified_before"},
	"navigation":          {"modified_after", "modified_before"},
	"pages":               {"modified_after", "modified_before"},
	"posts":               {"modified_after", "modified_before"},
}

var objectToResponseId = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
	"abilities":        "name",
	"block-types":      "name",
	"fields":           "name",
	"font-collections": "slug",
	"guideline-scopes": "slug",
	"icon-collections": "slug",
	"icons":            "name",
	"menu-locations":   "name",
	"plugins":          "name",
	"statuses":         "slug",
	"taxonomies":       "slug",
	"themes":           "name",
	"types":            "slug",
	"view-config":      "name",
}, func(objectName string) string {
	return "id"
})

func (c Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.MakeRecordsFunc(""), // Response has an array at root level, hence empty string.
		makeNextRecordsUrl(response),
		readhelper.MakeMarshaledSelectedDataFunc(
			fieldsSelector(params.ObjectName),
			jsonquery.Convertor.ObjectToMap,
		),
		params.Fields,
	)
}

func makeNextRecordsUrl(response *common.JSONHTTPResponse) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		return httpkit.HeaderLink(response, "next"), nil
	}
}

func fieldsSelector(objectName string) readhelper.SelectedFieldsFunc {
	return func(node *ajson.Node, fields []string) (map[string]any, string, error) {
		root, err := jsonquery.Convertor.ObjectToMap(node)
		if err != nil {
			return nil, "", err
		}

		identifierField := objectToResponseId.Get(objectName)

		identifier, err := jsonquery.New(node).TextWithDefault(identifierField, "")
		if err != nil {
			return nil, "", err
		}

		selected := readhelper.SelectFields(root, datautils.NewSetFromList(fields))

		return selected, identifier, nil
	}
}

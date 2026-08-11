package surveymonkey

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/surveymonkey/metadata"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v3"
	defaultPageSize = "100"

	pageKey    = "page"
	perPageKey = "per_page"
)

//nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectGroups,
	objectSurveyCategories,
	objectSurveyTemplates,
	objectTeamSurveyTemplates,
	objectSurveyLanguages,
	objectQuestionBankQuestions,
	objectSurveyFolders,
	objectContacts,
	objectContactLists,
	objectOrganizations,
	objectRoles,
	objectBenchmarkBundles,
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedReadObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	endpointURL, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	endpointURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, strings.TrimSpace(path))
	if err != nil {
		return nil, err
	}

	pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)

	endpointURL.WithQueryParam(pageKey, "1")
	endpointURL.WithQueryParam(perPageKey, pageSize)

	return endpointURL, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), params.ObjectName)

	return common.ParseResult(
		resp,
		extractRecordsOptional(recordsKey),
		nextPageFromLinks,
		readhelper.MakeMarshaledDataFuncWithId(nil, idFieldForObject(params.ObjectName)),
		params.Fields,
	)
}

func extractRecordsOptional(recordsKey string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(recordsKey)
	}
}

func nextPageFromLinks(node *ajson.Node) (string, error) {
	return jsonquery.New(node, "links").StrWithDefault("next", "")
}

func idFieldForObject(objectName string) readhelper.IdFieldQuery {
	switch objectName {
	case objectQuestionBankQuestions:
		return readhelper.NewIdField("question_id")
	case objectTeamSurveyTemplates:
		return readhelper.NewIdField("team_template_id")
	default:
		return readhelper.NewIdField("id")
	}
}

package breezy

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/breezy/metadata"
	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectCompanies,
	objectPositions,
	objectPipelines,
	objectCategories,
	objectDepartments,
	objectQuestionnaires,
	objectTemplates,
	objectWebhookEndpoints,
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedReadObjects.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	u, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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

	if params.ObjectName == "" {
		return nil, common.ErrMissingObjects
	}

	path, err := metadata.Schemas.FindURLPath(c.ProviderContext.Module(), params.ObjectName)
	if err != nil {
		return nil, err
	}

	if strings.Contains(path, companyIDPlaceholder) {
		if c.CompanyID == "" {
			return nil, ErrMissingCompanyID
		}

		path = resolveObjectPath(path, c.CompanyID)
	}

	endpointURL, err := buildVersionedPathURL(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, err
	}

	if params.ObjectName == objectPositions && params.Filter != "" {
		// Provider-side state filter (draft, archived, published, etc.).
		// https://developer.breezy.hr/reference/company-positions
		endpointURL.WithQueryParam("state", params.Filter)
	}

	return endpointURL, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResultFiltered(
		params,
		resp,
		nodeRecordsForRead(c, params.ObjectName),
		makeFilterFunc(params),
		readhelper.MakeMarshaledDataFuncWithId(nil, idFieldForObject(params.ObjectName)),
		params.Fields,
	)
}

func nodeRecordsForRead(c *Connector, objectName string) common.NodeRecordsFunc {
	switch objectName {
	case objectPipelines:
		return pipelineRecordNodes()
	default:
		recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName)

		return func(node *ajson.Node) ([]*ajson.Node, error) {
			return jsonquery.New(node).ArrayOptional(recordsKey)
		}
	}
}

func pipelineRecordNodes() common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		records, err := flattenPipelineRecords(node)
		if err != nil {
			return nil, err
		}

		return recordMapsToNodes(records)
	}
}

func flattenPipelineRecords(node *ajson.Node) ([]map[string]any, error) {
	m, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	out := make([]map[string]any, 0, len(m))

	for key, v := range m {
		obj, ok := v.(map[string]any)
		if !ok || obj == nil {
			continue
		}

		if _, has := obj["_id"]; !has && key != "" {
			obj["_id"] = key
		}

		out = append(out, obj)
	}

	return out, nil
}

func recordMapsToNodes(records []map[string]any) ([]*ajson.Node, error) {
	out := make([]*ajson.Node, 0, len(records))

	for _, record := range records {
		raw, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}

		node, err := ajson.Unmarshal(raw)
		if err != nil {
			return nil, err
		}

		out = append(out, node)
	}

	return out, nil
}

func idFieldForObject(objectName string) readhelper.IdFieldQuery {
	switch objectName {
	case objectCompanies, objectPositions, objectPipelines,
		objectDepartments, objectQuestionnaires, objectTemplates:
		return readhelper.NewIdField("_id")
	case objectCategories, objectWebhookEndpoints:
		return readhelper.NewIdField("id")
	default:
		return readhelper.NewIdField("id")
	}
}

func noNextPage(_ *ajson.Node) (string, error) {
	return "", nil
}

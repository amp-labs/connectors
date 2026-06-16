package breezy

import (
	"context"
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
		// Positions API defaults to state=published; Filter supplies draft, archived, etc.
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
	records, err := recordsForRead(c, params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		records,
		noNextPage,
		readhelper.MakeGetMarshaledDataWithId(idFieldForObject(params.ObjectName)),
		params.Fields,
	)
}

func recordsForRead(c *Connector, objectName string) (common.RecordsFunc, error) {
	switch objectName {
	case objectPipelines:
		// Pipelines payload is an object-of-objects (not a JSON array), so we must flatten it.
		// https://developer.breezy.hr/reference/company-pipelines
		return extractPipelinesRecords(), nil
	default:
		recordsKey := metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName)

		return common.ExtractOptionalRecordsFromPath(recordsKey), nil
	}
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

func extractPipelinesRecords() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
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
}

func noNextPage(_ *ajson.Node) (string, error) {
	return "", nil
}

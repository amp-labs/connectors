package breezy

import (
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/breezy/metadata"
	"github.com/spyzhov/ajson"
)

const (
	// Breezy API references:
	// - Companies: https://developer.breezy.hr/reference/companies
	// - Positions: https://developer.breezy.hr/reference/company-positions
	// - Webhook endpoints: https://developer.breezy.hr/reference/company-webhook-endpoints
	objectCompanies        = "companies"
	objectPositions        = "positions"
	objectWebhookEndpoints = "webhook_endpoints"
)

// nolint:gochecknoglobals
var supportedReadObjects = datautils.NewStringSet(
	objectCompanies,
	objectPositions,
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

	return endpointURL, nil
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		c.recordsFunc(params.ObjectName),
		noNextPage,
		readhelper.MakeMarshaledDataFuncWithId(nil, idFieldForObject(params.ObjectName)),
		params.Fields,
	)
}

func (c *Connector) recordsFunc(objectName string) common.NodeRecordsFunc {
	return common.MakeRecordsFunc(
		metadata.Schemas.LookupArrayFieldName(c.ProviderContext.Module(), objectName),
	)
}

func idFieldForObject(objectName string) readhelper.IdFieldQuery {
	switch objectName {
	case objectCompanies, objectPositions:
		return readhelper.NewIdField("_id")
	default:
		return readhelper.NewIdField("id")
	}
}

func noNextPage(_ *ajson.Node) (string, error) {
	return "", nil
}

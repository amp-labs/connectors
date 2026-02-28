package crm

import (
	"context"
	"log/slog"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

const data = "data"

var supportsIncSync = datautils.NewSet("activities", "deals", "organizations", "persons") // nolint: gochecknoglobals

func (a *Adapter) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	// Fetch custom field definitions so we can resolve hash keys to display names.
	// Graceful degradation: if this fails we proceed without custom field resolution.
	customFieldDefs, err := a.requestCustomFields(ctx, params.ObjectName)
	if err != nil {
		slog.Warn("Failed to fetch custom field definitions, continuing without custom field resolution",
			"object", params.ObjectName,
			"error", err)

		customFieldDefs = make(map[string]customFieldDef)
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.MakeRecordsFunc(data),
		nextRecordsURL(url),
		common.MakeMarshaledDataFunc(a.attachReadCustomFields(params.ObjectName, customFieldDefs)),
		params.Fields,
	)
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := a.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if supportsIncSync.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			url.WithQueryParam("updated_since", params.Since.Format(time.RFC3339))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("updated_until", params.Since.Format(time.RFC3339))
		}
	}

	return url, nil
}

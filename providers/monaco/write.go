package monaco

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/monaco/metadata"
)

// dataKey is where Monaco nests the written record in a write response:
// {"data": {...record...}, "meta": {...}}.
const dataKey = "data"

// writeConfig describes how one object is written. Monaco's write paths do not
// follow from its read paths -- reads go to /v1/contacts/list while writes go to
// /v1/contacts/ -- so they are spelled out rather than derived from schemas.json.
//
// Trailing slashes are reproduced exactly as the live API wants them; see url.go.
type writeConfig struct {
	// createPath is the collection endpoint used when RecordId is empty.
	// Empty means the object cannot be created.
	createPath string

	// createMethod is the verb for createPath. Normally POST.
	createMethod string

	// updatePath is the collection the RecordId is appended to, always via PATCH.
	// Empty means the object cannot be updated.
	updatePath string
}

//nolint:gochecknoglobals
var writeConfigs = map[string]writeConfig{
	objectAccounts: {
		// Accounts have no create endpoint -- POST /v1/accounts/ answers 405.
		// PUT is an upsert keyed on `domain`, which is the only way to insert
		// one, so a create is routed there. It will update an existing account
		// when the domain already exists rather than failing as a create would.
		createPath:   "/accounts/",
		createMethod: http.MethodPut,
		updatePath:   "/accounts",
	},
	objectAudiences: {
		createPath:   "/audiences",
		createMethod: http.MethodPost,
		// No update endpoint.
	},
	objectCampaigns: {
		createPath:   "/campaigns/",
		createMethod: http.MethodPost,
		updatePath:   "/campaigns",
	},
	objectContacts: {
		createPath:   "/contacts/",
		createMethod: http.MethodPost,
		updatePath:   "/contacts",
	},
	objectOpportunities: {
		createPath:   "/opportunities/",
		createMethod: http.MethodPost,
		updatePath:   "/opportunities",
	},
	objectSequenceTemplates: {
		createPath:   "/sequence-templates",
		createMethod: http.MethodPost,
		updatePath:   "/sequence-templates",
	},
	objectSequences: {
		// No create endpoint; sequences are produced by Monaco, not by callers.
		// PATCH takes an `action` rather than field values.
		updatePath: "/sequences",
	},
	objectTags: {
		createPath:   "/tags/",
		createMethod: http.MethodPost,
		updatePath:   "/tags",
	},
	objectTasks: {
		createPath:   "/tasks/",
		createMethod: http.MethodPost,
		updatePath:   "/tasks",
	},
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	endpointURL, method, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	// `id` is server-assigned and appears in no request schema. Dropping it
	// lets a record read back from Monaco be handed straight to Write.
	delete(record, "id")

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) buildWriteURL(params common.WriteParams) (string, string, error) {
	config, ok := writeConfigs[params.ObjectName]
	if !ok {
		return "", "", common.ErrOperationNotSupportedForObject
	}

	modulePath := metadata.Schemas.LookupModuleURLPath(c.ProviderContext.Module())

	if params.IsCreate() {
		if config.createPath == "" {
			return "", "", common.ErrOperationNotSupportedForObject
		}

		url, err := buildURL(c.ProviderInfo().BaseURL, modulePath, config.createPath)

		return url, config.createMethod, err
	}

	if config.updatePath == "" {
		return "", "", common.ErrOperationNotSupportedForObject
	}

	url, err := buildURL(c.ProviderInfo().BaseURL, modulePath, config.updatePath, params.RecordId)

	return url, http.MethodPatch, err
}

// parseWriteResponse reads the record back out of the `data` envelope. Monaco
// answers 200 for most writes and 201 for sequence template creation; both
// carry the same shape.
func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok || body == nil {
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	record, err := jsonquery.New(body).ObjectOptional(dataKey)
	if err != nil {
		return nil, err
	}

	if record == nil {
		// No `data` envelope: report success without inventing a record.
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	recordID, err := jsonquery.New(record).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(record)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

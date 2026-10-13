package jump

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/graphql"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// buildWriteRequest sends the record fields as variables of the object's create or update mutation.
// Variables are not rendered into the mutation text, so values are passed to Jump unchanged.
// Update mutations only change the fields provided.
// https://my.jumpapp.com/enterprise/documentation/reference?tab=mutations
func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	mutations, ok := writeMutations[params.ObjectName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	variables := maps.Clone(record)
	// Account-scoped API keys must name the user the mutation acts for.
	variables["userId"] = c.userID

	mutation := mutations.create

	if params.IsUpdate() {
		mutation = mutations.update
		variables["id"] = params.RecordId
	}

	return c.buildMutationRequest(ctx, mutation, variables)
}

func (c *Connector) buildMutationRequest(
	ctx context.Context,
	mutation string,
	variables map[string]any,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	query, err := graphql.Operation(queryFiles, "mutation", mutation, nil)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]any{
		"query":     query,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// parseWriteResponse returns the record located at "data.<mutation>".
//
//	{
//	  "data": {
//	    "createContact": {
//	      "id": "con_e0640d697bad4a2fb20",
//	      "name": "amp-test Contact",
//	      ...
//	    }
//	  }
//	}
func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	mutation := writeMutations[params.ObjectName].create
	if params.IsUpdate() {
		mutation = writeMutations[params.ObjectName].update
	}

	if err := interpretGraphQLError(resp, mutation); err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	record, err := jsonquery.New(body, "data").ObjectRequired(mutation)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(record).StringRequired("id")
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

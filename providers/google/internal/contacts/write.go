package contacts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/google/internal/core"
)

func (a *Adapter) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	operationName := core.OperationCreate
	if params.RecordId != "" {
		operationName = core.OperationUpdate
	}

	endpoint, err := endpoints.Find(operationName, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	url, err := a.getURL(endpoint.Path)
	if err != nil {
		return nil, err
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	if operationName == core.OperationUpdate && params.ObjectName == objectNameMyConnections {
		attachMyConnectionsWriteQueryParams(url, recordData)
	}

	recordData = nestPayload(params.ObjectName, recordData)

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, endpoint.Method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func nestPayload(objectName string, original map[string]any) map[string]any {
	if objectName != objectNameContactGroups {
		// Only "contactGroups" is wrapped in the name
		return original
	}

	singularObjectName := "contactGroup"
	if _, isWrapped := original[singularObjectName]; isWrapped {
		// User provided payload was already wrapped and is good to send as is.
		return original
	}

	// Payload should be nested with the name of the object to create/update.
	// Example payload to create contact group:
	// {"contactGroup": {"name": "coworkers"} }
	return map[string]any{
		singularObjectName: original,
	}
}

// Updating contact requires query parameter listing each field that is updated.
// https://developers.google.com/people/api/rest/v1/people/updateContact#query-parameters
func attachMyConnectionsWriteQueryParams(url *urlbuilder.URL, data map[string]any) {
	updateFields := datautils.FromMap(data).KeySet().Intersection(updatablePersonFields)
	url.WithQueryParam("updatePersonFields", strings.Join(updateFields, ","))
}

func (a *Adapter) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	resourceName, err := jsonquery.New(body).StringRequired("resourceName")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	recordID, idExists := resourceIdentifierFormat(resourceName)
	if !idExists {
		return &common.WriteResult{
			Success:  true,
			RecordId: "",
			Errors:   []any{common.ErrMissingRecordID},
			Data:     data,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

var updatablePersonFields = datautils.NewSet( // nolint:gochecknoglobals
	"addresses",
	"biographies",
	"birthdays",
	"calendarUrls",
	"clientData",
	"emailAddresses",
	"events",
	"externalIds",
	"genders",
	"imClients",
	"interests",
	"locales",
	"locations",
	"memberships",
	"miscKeywords",
	"names",
	"nicknames",
	"occupations",
	"organizations",
	"phoneNumbers",
	"relations",
	"sipAddresses",
	"urls",
	"userDefined",
)

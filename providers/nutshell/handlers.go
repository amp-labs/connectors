package nutshell

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/nutshell/internal/metadata"
	"github.com/spyzhov/ajson"
)

// The API accepts large page numbers without an error.
// The actual maximum limit is 10,000; anything higher will return an error.
const defaultPageSize = "10000"

// Events: https://developers.nutshell.com/reference/get_events
const objectNameEvents = "events"

// https://developers.nutshell.com/reference/post_notes
const objectNameNotes = "notes"

var updateHeader = common.Header{ // nolint:gochecknoglobals
	Key:   "Content-Type",
	Value: "application/json-patch+json",
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// The Events endpoint uses parameter names that differ from other endpoints.
	if params.ObjectName == objectNameEvents {
		url.WithQueryParam("limit", defaultPageSize)
	} else {
		url.WithQueryParam("page[limit]", defaultPageSize)
	}

	// Time queries need a range.
	// Giving an Until without a Since is ignored.
	// Passing only Since will make Until default to now.
	// https://developers.nutshell.com/docs/filters#data-filters
	if incrementalReadByCreatedTimeQP.Has(params.ObjectName) && !params.Since.IsZero() {
		timeUntil := params.Until
		if timeUntil.IsZero() {
			timeUntil = time.Now()
		}

		url.WithQueryParam("filter[createdTime]", fmt.Sprintf("%v %v",
			datautils.Time.FormatRFC3339inUTC(params.Since),
			datautils.Time.FormatRFC3339inUTC(timeUntil),
		))
	}

	// The Events endpoint uses parameter names that differ from other endpoints.
	if params.ObjectName == objectNameEvents {
		if !params.Since.IsZero() {
			url.WithQueryParam("since_time", datautils.Time.Unix(params.Since))
		}

		if !params.Until.IsZero() {
			url.WithQueryParam("max_time", datautils.Time.Unix(params.Until))
		}
	}

	return url, nil
}

var incrementalReadByCreatedTimeQP = datautils.NewSet( // nolint:gochecknoglobals
	"accounts",
	"activities",
	"contacts",
	"leads",
)

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module(), params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	// Alter current request URL to progress with the next page token.
	return func(node *ajson.Node) (string, error) {
		// This response structure relates to `events,notes` objects.
		metaObject, err := jsonquery.New(node).ObjectOptional("meta")
		if err != nil {
			return "", err
		}

		if metaObject != nil && metaObject.HasKey("next") {
			// Events, notes objects use this format where `meta` holds `next` page token.
			return jsonquery.New(metaObject).StrWithDefault("next", "")
		}

		page, exists := url.GetFirstQueryParam("page[page]")
		if !exists {
			// First page count begins from zero. Current page is zeroth page.
			page = "0"
		}

		pageNum, err := strconv.Atoi(page)
		if err != nil {
			return "", err
		}

		// Advance pagination counter.
		pageNum += 1

		url.WithQueryParam("page[page]", strconv.Itoa(pageNum))

		return url.String(), nil
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	if len(params.RecordId) == 0 {
		return c.buildCreateRequest(ctx, params, url)
	}

	return c.buildUpdateRequest(ctx, params, url)
}

func (c *Connector) buildCreateRequest(
	ctx context.Context, params common.WriteParams, url *urlbuilder.URL,
) (*http.Request, error) {
	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	// Create payload must be wrapped in the object name.
	/* {
	  "sources": [
	    {
	      "name": "NorthWest",
	      "channel": 3
	    }
	  ]
	} */
	payloadWrapperKey := payloadWrapperKeyRegistry.Get(params.ObjectName)
	if _, correctlyFormatted := recordData[payloadWrapperKey]; !correctlyFormatted {
		var value any = []any{recordData}

		if params.ObjectName == objectNameNotes {
			/* {
			  "data": {
			    "links": {
			      "parent": "object-id-which-is-target-for-the-note"
			    },
			    "body": "Note content goes here"
			  }
			} */
			value = recordData
		}

		recordData = map[string]any{
			payloadWrapperKey: value,
		}
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *Connector) buildUpdateRequest(
	ctx context.Context, params common.WriteParams, url *urlbuilder.URL,
) (*http.Request, error) {
	// Operations must be provided as an array.
	// Since params.RecordData cannot be an array itself,
	// we wrap it in a single-element array.
	recordData := []any{
		params.RecordData,
	}

	jsonData, err := json.Marshal(recordData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	updateHeader.ApplyToRequest(req)

	return req, nil
}

// Creating objects requires payload to be wrapped with some key.
// The object name doesn't always match the payload key, the registry bellow handles this.
var payloadWrapperKeyRegistry = datautils.NewDefaultMap(map[string]string{ // nolint:gochecknoglobals
	"audiences": "emAudiences",
	"notes":     "data",
}, func(objectName string) (payloadKey string) {
	return objectName
})

func (c *Connector) parseWriteResponse(ctx context.Context, params common.WriteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	array, err := jsonquery.New(body).ArrayRequired(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(array) == 0 {
		// Response doesn't hold any data for the object.
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	nested := array[0]

	recordID, err := jsonquery.New(nested).TextWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(nested)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := c.getWriteURL(params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(ctx context.Context, params common.DeleteParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if response.Code != http.StatusOK && response.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, response.Code)
	}

	// Response body is not used.
	return &common.DeleteResult{
		Success: true,
	}, nil
}

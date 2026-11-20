package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// HubSpot returns 477 while a hub is being migrated between data hosting locations.
// Migrations resolve in hours, so it is a retryable error.
const hubspotMigrationStatusCode = 477

// invalidPaginationCursorMessageSubstring is the message HubSpot returns when its
// search/list APIs reject the supplied pagination cursor as unparsable.
// This is caused by the pagination cursor being not the appropriate type
// for the API (because Search and List APIs use different pagination tokens).
const invalidPaginationCursorMessageSubstring = "Cannot deserialize value of type"

// InterpretJSONError interprets the error response from Hubspot
// as per https://developers.hubspot.com/docs/api/error-handling.
func InterpretJSONError(res *http.Response, body []byte) error {
	headers := common.GetResponseHeaders(res)

	apiError := &HubspotError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return common.NewHTTPError(res.StatusCode, body, headers, errors.Join(common.ErrFailedUnmarshalling, err))
	}

	switch res.StatusCode {
	case http.StatusBadRequest:
		if strings.Contains(apiError.Message, invalidPaginationCursorMessageSubstring) {
			return common.NewHTTPError(res.StatusCode, body, headers,
				createError(common.ErrInvalidPaginationCursor, apiError))
		}
		// Hubspot sends us a 400 when the search endpoint returns over 10K records,
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrBadRequest, apiError))
	case http.StatusUnauthorized:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrAccessToken, apiError))
	case http.StatusForbidden:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrForbidden, apiError))
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusGatewayTimeout:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrLimitExceeded, apiError))
	case http.StatusServiceUnavailable:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrApiDisabled, apiError))
	case hubspotMigrationStatusCode:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrRetryable, apiError))
	default:
		return common.InterpretError(res, body)
	}
}

func createError(baseErr error, hubspotError *HubspotError) error {
	if len(hubspotError.Message) > 0 {
		return fmt.Errorf("%w: %s: %+v", baseErr, hubspotError.Message, hubspotError.Details)
	}

	return baseErr
}

type HubspotError struct {
	HTTPStatusCode int         `json:"httpStatusCode"`
	Status         string      `json:"status,omitempty"`
	Message        string      `json:"message,omitempty"`
	CorrelationID  string      `json:"correlationId,omitempty"`
	Context        ErrContext  `json:"context"`
	Category       string      `json:"category,omitempty"`
	SubCategory    string      `json:"subCategory,omitempty"`
	Links          ErrLinks    `json:"links"`
	Details        []ErrDetail `json:"details,omitempty"`
}

type ErrDetail struct {
	IsValid bool   `json:"isValid,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Name    string `json:"name,omitempty"`
}

type ErrContext struct {
	ID             []string `json:"id,omitempty"`
	Type           []string `json:"type,omitempty"`
	ObjectType     []string `json:"objectType,omitempty"`
	FromObjectType []string `json:"fromObjectType,omitempty"`
	ToObjectType   []string `json:"toObjectType,omitempty"`
}

type ErrLinks struct {
	APIKey        string `json:"apiKey,omitempty"`
	KnowledgeBase string `json:"knowledgeBase,omitempty"`
}

// Error categories HubSpot reports in the errors array of a batch response.
const (
	// Seen in a captured batch update response: test/batch/update/contacts/err-partial-success.json.
	batchErrorCategoryObjectNotFound = "OBJECT_NOT_FOUND"
	// https://developers.hubspot.com/docs/api/error-handling
	batchErrorCategoryValidation = "VALIDATION_ERROR"
)

// errBatchInputFailed is used for a batch error whose category has no matching common error.
var errBatchInputFailed = errors.New("hubspot batch input failed")

// Err converts the batch error to an error that callers can classify: a missing object wraps
// common.ErrNotFound and a validation error wraps common.ErrBadRequest.
func (e BatchError) Err() error {
	switch e.Category {
	case batchErrorCategoryObjectNotFound:
		return fmt.Errorf("%w: %s", common.ErrNotFound, e.Message)
	case batchErrorCategoryValidation:
		return fmt.Errorf("%w: %s", common.ErrBadRequest, e.Message)
	default:
		return fmt.Errorf("%w: %s: %s", errBatchInputFailed, e.Category, e.Message)
	}
}

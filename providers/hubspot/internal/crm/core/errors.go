package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

// InterpretJSONError interprets the error response from Hubspot
// as per https://developers.hubspot.com/docs/api/error-handling.
func InterpretJSONError(res *http.Response, body []byte) error {
	apiError := &HubspotError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return fmt.Errorf("status code %v and json.Unmarshal failed: %w", res.StatusCode, err)
	}

	headers := common.GetResponseHeaders(res)

	switch res.StatusCode {
	// Hubspot sends us a 400 when the search endpoint returns over 10K records.
	case http.StatusBadRequest:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrBadRequest, apiError))
	case http.StatusUnauthorized:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrAccessToken, apiError))
	case http.StatusForbidden:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrForbidden, apiError))
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusGatewayTimeout:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrLimitExceeded, apiError))
	case http.StatusServiceUnavailable:
		return common.NewHTTPError(res.StatusCode, body, headers, createError(common.ErrApiDisabled, apiError))
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

package hubspot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrMissingAPIModule = errors.New("missing Hubspot API module")
	ErrMissingClient    = errors.New("JSON http client not set")
	ErrNotArray         = errors.New("results is not an array")
	ErrNotObject        = errors.New("result is not an object")
	ErrNotString        = errors.New("link is not a string")
)

type HubspotError struct {
	HTTPStatusCode int         `json:"httpStatusCode"`
	Status         string      `json:"status,omitempty"`
	Message        string      `json:"message,omitempty"`
	CorrelationID  string      `json:"correlationId,omitempty"`
	Context        ErrContext  `json:"context,omitempty"`
	Category       string      `json:"category,omitempty"`
	SubCategory    string      `json:"subCategory,omitempty"`
	Links          ErrLinks    `json:"links,omitempty"`
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

func (c *Connector) interpretError(res *http.Response, body []byte) error {
	if res.Header.Get("Content-Type") == "application/json" {
		return c.interpretJSONError(res, body)
	}

	return common.InterpretError(res, body)
}

func createError(baseErr error, hubspotError *HubspotError) error {
	if len(hubspotError.Message) > 0 {
		return fmt.Errorf("%w: %s: %+v", baseErr, hubspotError.Message, hubspotError.Details)
	}

	return baseErr
}

// interpretJSONError interprets the error response from Hubspot
// as per https://developers.hubspot.com/docs/api/error-handling.
func (c *Connector) interpretJSONError(res *http.Response, body []byte) error {
	apiError := &HubspotError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return createError(common.ErrAccessToken, apiError)
	case http.StatusForbidden:
		return createError(common.ErrForbidden, apiError)
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusGatewayTimeout:
		return createError(common.ErrLimitExceeded, apiError)
	case http.StatusServiceUnavailable:
		return createError(common.ErrApiDisabled, apiError)
	default:
		return common.InterpretError(res, body)
	}
}

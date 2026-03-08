package restlet

import (
	"encoding/json"

	"github.com/amp-labs/connectors/common"
)

const statusSuccess = "SUCCESS"

// lookupFilterOperator maps an Ampersand filter operator to a NetSuite search operator.
func lookupFilterOperator(op common.FilterOperator) (string, bool) {
	switch op {
	case common.FilterOperatorEQ:
		return "is", true
	default:
		return "", false
	}
}

// restletResponse is the envelope returned by every RESTlet action.
type restletResponse struct {
	Header responseHeader  `json:"header"`
	Body   json.RawMessage `json:"body"`
}

// responseHeader contains status and pagination metadata.
type responseHeader struct {
	Status       string `json:"status"`
	NextPage     *int   `json:"nextPage"`
	HasMore      bool   `json:"hasMore"`
	TotalResults int64  `json:"totalResults"`
	TotalPages   int    `json:"totalPages"`
	PageSize     int    `json:"pageSize"`
	Version      string `json:"version"`
}

// restletErrorBody is the body when header.status == "ERROR".
type restletErrorBody struct {
	ErrorCode    string          `json:"errorCode"`
	ErrorMessage string          `json:"errorMessage"`
	ErrorDetails json.RawMessage `json:"errorDetails,omitempty"`
}

// searchRequest is the JSON body for the search action.
type searchRequest struct {
	Action    string     `json:"action"`
	Type      string     `json:"type"`
	Columns   []string   `json:"columns,omitempty"`
	Filters   []any      `json:"filters,omitempty"`
	Sort      []sortSpec `json:"sort,omitempty"`
	PageSize  int        `json:"pageSize,omitempty"`
	PageIndex int        `json:"pageIndex,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

type sortSpec struct {
	Column    string `json:"column"`
	Direction string `json:"direction"`
}

// writeRequest is the JSON body for create/update actions.
type writeRequest struct {
	Action   string         `json:"action"`
	Type     string         `json:"type"`
	RecordId string         `json:"recordId,omitempty"`
	Values   map[string]any `json:"values,omitempty"`
	Sublists map[string]any `json:"sublists,omitempty"`
}

// deleteRequest is the JSON body for the delete action.
type deleteRequest struct {
	Action   string `json:"action"`
	Type     string `json:"type"`
	RecordId string `json:"recordId"`
}

// schemaRequest is the JSON body for the getschema action.
type schemaRequest struct {
	Action string `json:"action"`
	Type   string `json:"type"`
}

// writeResponseBody is the body on a successful create/update.
type writeResponseBody struct {
	RecordId any    `json:"recordId"` // can be int or string
	Type     string `json:"type"`
}

// schemaResponseBody is the body on a successful getschema.
type schemaResponseBody struct {
	Type         string                       `json:"type"`
	SchemaSource string                       `json:"schemaSource"`
	Fields       map[string]schemaFieldInfo   `json:"fields"`
	Sublists     map[string]schemaSublistInfo `json:"sublists"`
}

type schemaFieldInfo struct {
	Label       string `json:"label"`
	Type        string `json:"type"`
	IsMandatory bool   `json:"isMandatory"`
	IsReadOnly  bool   `json:"isReadOnly"`
}

type schemaSublistInfo struct {
	Fields map[string]schemaFieldInfo `json:"fields"`
}

// searchResultRow is a single record in the search response body.
// Fields are dynamic: _id, _type, and then each column as {value, text}.
type searchFieldValue struct {
	Value any    `json:"value"`
	Text  string `json:"text"`
}

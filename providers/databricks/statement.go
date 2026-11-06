package databricks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	statementsPath = "api/2.0/sql/statements"

	// The API waits at most 50 seconds for a result. A statement that is still
	// running after that is canceled, so every request is a single round trip.
	statementWaitTimeout = "50s"

	stateSucceeded = "SUCCEEDED"

	// Databricks timestamp parameters accept ISO 8601 with an explicit zone.
	timestampParameterLayout = "2006-01-02T15:04:05.000000Z"
)

type statementRequest struct {
	WarehouseID   string               `json:"warehouse_id"`
	Statement     string               `json:"statement"`
	Parameters    []statementParameter `json:"parameters,omitempty"`
	WaitTimeout   string               `json:"wait_timeout"`
	OnWaitTimeout string               `json:"on_wait_timeout"`
	Format        string               `json:"format"`
	Disposition   string               `json:"disposition"`
}

type statementParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

type statementResponse struct {
	StatementID string `json:"statement_id"`
	Status      struct {
		State string `json:"state"`
		Error *struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		} `json:"error"`
	} `json:"status"`
	Manifest *struct {
		Schema struct {
			Columns []statementColumn `json:"columns"`
		} `json:"schema"`
	} `json:"manifest"`
	Result *struct {
		DataArray [][]any `json:"data_array"`
	} `json:"result"`
}

type statementColumn struct {
	Name     string `json:"name"`
	TypeName string `json:"type_name"`
}

func (c *Connector) newStatementRequest(statement string, params ...statementParameter) statementRequest {
	return statementRequest{
		WarehouseID:   c.warehouseID,
		Statement:     statement,
		Parameters:    params,
		WaitTimeout:   statementWaitTimeout,
		OnWaitTimeout: "CANCEL",
		Format:        "JSON_ARRAY",
		Disposition:   "INLINE",
	}
}

func (c *Connector) statementsURL() (string, error) {
	u, err := urlbuilder.New(c.ProviderInfo().BaseURL, statementsPath)
	if err != nil {
		return "", fmt.Errorf("build databricks statements url: %w", err)
	}

	return u.String(), nil
}

func (c *Connector) buildStatementHTTPRequest(
	ctx context.Context, body statementRequest,
) (*http.Request, error) {
	urlStr, err := c.statementsURL()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal databricks statement: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Connector) executeStatement(ctx context.Context, body statementRequest) ([]map[string]any, error) {
	urlStr, err := c.statementsURL()
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Post(ctx, urlStr, body)
	if err != nil {
		return nil, err
	}

	return parseStatementRecords(resp)
}

// parseStatementRecords turns a statement response into one map per row, keyed by
// column name. An HTTP 200 can still carry a failed statement, so the state is checked.
func parseStatementRecords(resp *common.JSONHTTPResponse) ([]map[string]any, error) {
	parsed, err := common.UnmarshalJSON[statementResponse](resp)
	if err != nil {
		return nil, err
	}

	if parsed.Status.State != stateSucceeded {
		message := ""
		if parsed.Status.Error != nil {
			message = parsed.Status.Error.ErrorCode + ": " + parsed.Status.Error.Message
		}

		return nil, fmt.Errorf("%w: state %s %s", errStatementNotSucceeded, parsed.Status.State, message)
	}

	if parsed.Manifest == nil || parsed.Result == nil {
		return []map[string]any{}, nil
	}

	columns := parsed.Manifest.Schema.Columns
	records := make([]map[string]any, 0, len(parsed.Result.DataArray))

	for _, row := range parsed.Result.DataArray {
		record := make(map[string]any, len(columns))

		for index, column := range columns {
			if index >= len(row) {
				break
			}

			record[column.Name] = convertCell(column.TypeName, row[index])
		}

		records = append(records, record)
	}

	return records, nil
}

// convertCell restores the column type. JSON_ARRAY format sends every non-null value
// as a string.
func convertCell(typeName string, value any) any {
	text, isString := value.(string)
	if !isString {
		return value
	}

	switch strings.ToUpper(typeName) {
	case "BYTE", "SHORT", "INT", "LONG":
		if number, err := strconv.ParseInt(text, 10, 64); err == nil {
			return number
		}
	case "FLOAT", "DOUBLE", "DECIMAL":
		if number, err := strconv.ParseFloat(text, 64); err == nil {
			return number
		}
	case "BOOLEAN":
		if flag, err := strconv.ParseBool(text); err == nil {
			return flag
		}
	case "ARRAY", "MAP", "STRUCT", "VARIANT":
		var nested any
		if err := json.Unmarshal([]byte(text), &nested); err == nil {
			return nested
		}
	}

	return text
}

// quoteIdentifier wraps a catalog, schema, table or column name in backticks.
func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func timestampParameter(name string, value time.Time) statementParameter {
	return statementParameter{
		Name:  name,
		Value: value.UTC().Format(timestampParameterLayout),
		Type:  "TIMESTAMP",
	}
}

package outplay

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func inferValueTypeFromData(value any) common.ValueType {
	if value == nil {
		return common.ValueTypeOther
	}

	switch value.(type) {
	case string:
		return common.ValueTypeString
	case float64, int, int64:
		return common.ValueTypeFloat
	case bool:
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func buildMetadataBody(objectName string) (*bytes.Reader, error) {
	body := map[string]any{
		"pageindex": 1,
	}

	if objectName == ObjectNameCall {
		now := time.Now()

		// Call object requires fromdate and todate parameters.
		// We set todate as current date and fromdate as 30 days ago to get recent calls.
		thirtyDaysAgo := now.AddDate(0, 0, -30)

		body["fromdate"] = thirtyDaysAgo.Format(timeLayout)
		body["todate"] = now.Format(timeLayout)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(bodyJSON)

	return bodyReader, nil
}

func buildReadBody(params common.ReadParams) (*bytes.Reader, error) {
	body := map[string]any{}

	if params.NextPage != "" {
		pageIndex, err := strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		body["pageindex"] = pageIndex
	} else {
		body["pageindex"] = 1
	}

	// call object requires date filters
	if params.ObjectName == ObjectNameCall {
		// There is no limit on date range but it is mandatory to provide fromdate and todate
		// we set default date range as last 30 days
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()

		if !params.Since.IsZero() {
			startDate = params.Since
		}

		if !params.Until.IsZero() {
			endDate = params.Until
		}

		body["fromdate"] = startDate.Format(timeLayout)
		body["todate"] = endDate.Format(timeLayout)
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(bodyJSON), nil
}

func buildReadQueryParams(url *urlbuilder.URL, params common.ReadParams) {
	if !params.Since.IsZero() {
		url.WithQueryParam("fromdate", params.Since.Format(timeLayout))
	}

	if !params.Until.IsZero() {
		url.WithQueryParam("todate", params.Until.Format(timeLayout))
	}

	if params.NextPage != "" {
		url.WithQueryParam("pageindex", params.NextPage.String())
	} else {
		url.WithQueryParam("pageindex", "1")
	}
}

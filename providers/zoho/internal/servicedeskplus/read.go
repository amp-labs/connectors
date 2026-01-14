package servicedeskplus

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

const (
	pageSize     = 100
	timestampKey = "display_value"
	timeLayout   = "Jan 2, 2006 03:04 PM"
)

type listInfo struct {
	RowCount       int      `json:"row_count,omitempty"`
	StartIndex     int      `json:"start_index,omitempty"`
	SortField      string   `json:"sort_field,omitempty"`
	SortOrder      string   `json:"sort_order,omitempty"`
	Page           int      `json:"page,omitempty"`
	GetTotalCount  bool     `json:"get_total_count,omitempty"`
	FieldsRequired []string `json:"fields_required,omitempty"`
}

type InputData struct {
	ListInfo *listInfo `json:"list_info"`
}

func (input InputData) MarshalToString() (string, error) {
	listInfoJSON, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("error marshaling list_info: %w", err)
	}

	return string(listInfoJSON), nil
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	url, err := a.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	input := InputData{
		&listInfo{
			RowCount: pageSize,
			// some of the APIs have updated_time field, but
			// they mostly are nil for non updated records.
			GetTotalCount:  true,
			FieldsRequired: params.Fields.List(),
		},
	}

	if hasCreationDate.Has(params.ObjectName) {
		input.ListInfo.SortField = "created_date"
		input.ListInfo.SortOrder = "desc"
	} else if hasCreationTime.Has(params.ObjectName) {
		input.ListInfo.SortField = "created_time"
		input.ListInfo.SortOrder = "desc"
	}

	if params.NextPage != "" {
		nextPage, err := strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		input.ListInfo.Page = nextPage
	}

	inf, err := input.MarshalToString()
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("input_data", inf)

	return url, nil
}

func (a *Adapter) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	h := common.Header{
		Key:   "Accept",
		Value: "application/vnd.manageengine.sdp.v3+json",
	}

	res, err := a.Client.Get(ctx, url.String(), h)
	if err != nil {
		return nil, err
	}

	if !config.Since.IsZero() {
		node, hasData := res.Body()
		if !hasData {
			return &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			}, nil
		}

		return manualIncrementalSync(node, config.ObjectName, config, timestampKey, timeLayout, getNextRecordsURL)
	}

	return common.ParseResult(res,
		extractRecordsFromPath(config.ObjectName),
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}

func manualIncrementalSync(node *ajson.Node, recordsKey string, config common.ReadParams, //nolint:cyclop
	timestampKey string, timestampFormat string, nextPageFunc common.NextPageFunc,
) (*common.ReadResult, error) {
	var zoomField string
	if hasCreationDate.Has(config.ObjectName) {
		zoomField = "created_date"
	}

	if hasCreationTime.Has(config.ObjectName) {
		zoomField = "created_time"
	}

	records, nextPage, err := readhelper.FilterSortedRecords(node, recordsKey,
		config.Since, timestampKey, timestampFormat, nextPageFunc, zoomField)
	if err != nil {
		return nil, err
	}

	rows, err := common.GetMarshaledData(records, config.Fields.List())
	if err != nil {
		return nil, err
	}

	var done bool
	if nextPage == "" {
		done = true
	}

	return &common.ReadResult{
		Rows:     int64(len(records)),
		Data:     rows,
		NextPage: common.NextPageToken(nextPage),
		Done:     done,
	}, nil
}

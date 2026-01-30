package apollo

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// Search uses POST method to read data.It has a display limit of 50,000 records.
// It's recommended to filter the results so as to narrow down the results as much as possible.
// Most of the Filtering would need client's input so we don't exhaust calls by paging through all 50k records.
// Using this as is may lead to that issue.
func (c *Connector) Search(ctx context.Context, config common.ReadParams, url *urlbuilder.URL,
) (*common.ReadResult, error) {
	resp, err := c.Client.Post(ctx, url.String(), []byte{})
	if err != nil {
		return nil, err
	}

	node, ok := resp.Body()
	if !ok {
		return nil, err
	}

	recordsFieldKey := constructSupportedObjectName(config.ObjectName)

	if !config.Since.IsZero() {
		if config.ObjectName == contacts {
			return manualIncrementalSync(node, recordsFieldKey, config, updatedAt, time.RFC3339, getNextRecords)
		}

		// We cannot filter records by the updatedAt field because the API response does not include it,
		// even though sorting by it is supported. So we currently use the createdAt field.
		if config.ObjectName == accounts {
			return manualIncrementalSync(node, recordsFieldKey, config, createdAt, time.RFC3339, getNextRecords)
		}
	}

	return common.ParseResult(
		resp,
		searchRecords(config.ObjectName),
		getNextRecords,
		apolloMarshaledData(config.ObjectName),
		config.Fields,
	)
}

// Manual incremental synchronization implementation for Apollo
//
// Apollo lacks native incremental sync support. This function iterates through records
// and returns those created or updated after the specified timestamp.
func manualIncrementalSync(node *ajson.Node, recordsKey string, config common.ReadParams, //nolint:cyclop
	timestampKey string, timestampFormat string, nextPageFunc common.NextPageFunc,
) (*common.ReadResult, error) {
	records, nextPage, err := readhelper.FilterSortedRecords(node, recordsKey,
		config.Since, timestampKey, timestampFormat, nextPageFunc)
	if err != nil {
		return nil, err
	}

	rows, err := getMarshaledData(records, config.Fields.List())
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

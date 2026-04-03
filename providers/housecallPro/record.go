package housecallpro

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/housecallPro/metadata"
)

func (c *Connector) urlPathForRecordByID(objectName string) (string, error) {
	if objectName == "invoices" {
		return "/api/invoices", nil
	}

	return metadata.Schemas.FindURLPath(c.Module(), objectName)
}

// GetRecordsByIds implements connectors.BatchRecordReaderConnector.
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	recordIDs []string,
	fields []string,
	associations []string,
) ([]common.ReadResultRow, error) {
	path, err := c.urlPathForRecordByID(objectName)
	if err != nil {
		return nil, err
	}

	marshal := readhelper.MakeGetMarshaledDataWithId(readIDFieldByObject.Get(objectName))
	out := make([]common.ReadResultRow, 0, len(recordIDs))

	// Single object returned per webhook
	for _, recordID := range recordIDs {
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
		if err != nil {
			return nil, err
		}

		url.AddPath(recordID)

		resp, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		body, ok := resp.Body()
		if !ok {
			return nil, common.ErrEmptyJSONHTTPResponse
		}

		recordMap, err := jsonquery.Convertor.ObjectToMap(body)
		if err != nil {
			return nil, err
		}

		rows, err := marshal([]map[string]any{recordMap}, fields)
		if err != nil {
			return nil, err
		}

		out = append(out, rows...)
	}

	return out, nil
}

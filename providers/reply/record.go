package reply

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

var _ connectors.BatchRecordReaderConnector = &Connector{}

// recordObjects lists the objects that have a get-by-id endpoint: GET /v3/{object}/{id}.
// https://docs.reply.io/api-reference/contacts/get-a-contact
//
//nolint:gochecknoglobals
var recordObjects = datautils.NewStringSet(
	objectContacts,
	"contact-accounts",
	objectSequences,
	objectEmailAccounts,
	objectLinkedInAccount,
)

// GetRecordsByIds fetches full records from Reply for a specific set of IDs. Reply has no
// batch endpoint, so it makes one GET per id, in order, which also keeps a burst of webhook
// events within the rate limit (100 requests per minute).
//
// An id that no longer exists is skipped rather than failing the batch, so the caller
// receives the records that do resolve. Each row is built the way Read builds it.
func (c *Connector) GetRecordsByIds( //nolint:revive
	ctx context.Context,
	objectName string,
	recordIds []string,
	fields []string,
	_ []string,
) ([]common.ReadResultRow, error) {
	objectName = strings.ToLower(objectName)

	if !recordObjects.Has(objectName) {
		return nil, fmt.Errorf("%w: %s", common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return []common.ReadResultRow{}, nil
	}

	marshal := readhelper.MakeMarshaledDataFuncWithId(nil, readhelper.NewIdField("id"))
	rows := make([]common.ReadResultRow, 0, len(recordIds))

	for _, recordID := range recordIds {
		record, err := c.fetchRecord(ctx, objectName, recordID)
		if err != nil {
			if isNotFound(err) {
				continue
			}

			return nil, fmt.Errorf("fetch %s/%s: %w", objectName, recordID, err)
		}

		marshaled, err := marshal([]*ajson.Node{record}, fields)
		if err != nil {
			return nil, fmt.Errorf("parse %s/%s: %w", objectName, recordID, err)
		}

		rows = append(rows, marshaled...)
	}

	return rows, nil
}

func (c *Connector) fetchRecord(ctx context.Context, objectName, recordID string) (*ajson.Node, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName, recordID)
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return body, nil
}

// isNotFound reports whether err is a 404 from Reply. The base JSON HTTP client does not
// always map a 404 to common.ErrNotFound, so the HTTP status is inspected as well.
func isNotFound(err error) bool {
	if errors.Is(err, common.ErrNotFound) {
		return true
	}

	if httpErr, ok := errors.AsType[*common.HTTPError](err); ok {
		return httpErr.Status == http.StatusNotFound
	}

	return false
}

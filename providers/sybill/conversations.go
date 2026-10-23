package sybill

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/sybill/internal/metadata"
	"github.com/spyzhov/ajson"
)

// Fields that only GET /v1/conversations/{conversationId} returns, in lowercase.
// https://api.sybill.ai/docs/data-models.html#conversation-detail
var conversationDetailFields = datautils.NewSet( // nolint:gochecknoglobals
	"category",
	"summary",
	"transcript",
	"recordings",
)

// needsConversationDetails reports whether a conversations read asks for a detail-only field.
// Reads of list fields alone make no detail requests.
func needsConversationDetails(params common.ReadParams) bool {
	if params.ObjectName != objectNameConversations {
		return false
	}

	for _, field := range params.Fields.List() {
		if conversationDetailFields.Has(strings.ToLower(field)) {
			return true
		}
	}

	return false
}

type conversationsPage struct {
	Conversations []struct {
		ConversationID string `json:"conversationId"`
	} `json:"conversations"`
}

// fetchConversationDetails gets the detail of every conversation on the page, one at a time:
// the per-key limit of 60 requests a minute leaves nothing to gain from concurrency.
func (c *Connector) fetchConversationDetails(
	ctx context.Context, response *common.JSONHTTPResponse,
) (map[string]map[string]any, error) {
	page, err := common.UnmarshalJSON[conversationsPage](response)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	details := make(map[string]map[string]any, len(page.Conversations))

	for _, conversation := range page.Conversations {
		detail, err := c.fetchConversationDetail(ctx, conversation.ConversationID)
		if err != nil {
			return nil, err
		}

		details[conversation.ConversationID] = detail
	}

	return details, nil
}

func (c *Connector) fetchConversationDetail(ctx context.Context, conversationID string) (map[string]any, error) {
	path, err := metadata.Catalog.ReadPath(objectNameConversations)
	if err != nil {
		return nil, err
	}

	detailURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path, conversationID)
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Get(ctx, detailURL.String())
	if err != nil {
		return nil, err
	}

	detail, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, err
	}

	if detail == nil {
		return nil, fmt.Errorf("%w: conversation %s", common.ErrEmptyJSONHTTPResponse, conversationID)
	}

	return *detail, nil
}

// marshalConversationDetails returns each listed conversation as its detail, which holds every list
// field plus transcript, summary, recordings and category. Raw is the detail response.
func marshalConversationDetails(details map[string]map[string]any) common.MarshalFromNodeFunc {
	return func(records []*ajson.Node, fields []string) ([]common.ReadResultRow, error) {
		rows := make([]common.ReadResultRow, len(records))

		for index, record := range records {
			conversationID, err := jsonquery.New(record).StringRequired(idFields[objectNameConversations])
			if err != nil {
				return nil, err
			}

			detail, ok := details[conversationID]
			if !ok {
				return nil, fmt.Errorf("%w: detail of conversation %s", common.ErrMissingExpectedValues, conversationID)
			}

			rows[index] = common.ReadResultRow{
				Fields: common.ExtractLowercaseFieldsFromRaw(fields, detail),
				Raw:    detail,
				Id:     conversationID,
			}
		}

		return rows, nil
	}
}

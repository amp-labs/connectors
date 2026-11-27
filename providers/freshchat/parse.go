package freshchat

import (
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	objectAccounts = "accounts"
	objectAgents   = "agents"
	objectUsers    = "users"

	objectConversations    = "conversations"
	objectOutboundWhatsApp = "outbound-messages/whatsapp"

	defaultIDField = "id"

	// agentStatusField is the only property of an agent that is changed with PATCH.
	agentStatusField = "agent_status"
)

var supportFilteringByTime = datautils.NewStringSet(objectUsers) //nolint: gochecknoglobals

// readPaths lists the objects which are not read from the path named after them.
var readPaths = datautils.Map[string, string]{ //nolint: gochecknoglobals
	objectAccounts: "accounts/configuration",
}

// writeIDFields lists the objects whose identifier is found under a name other than `id`
// in the answer to a write.
var writeIDFields = datautils.Map[string, string]{ //nolint: gochecknoglobals
	objectConversations:    "conversation_id",
	objectOutboundWhatsApp: "request_id",
}

// singleRecordObjects are answered with the record itself rather than a list of records.
var singleRecordObjects = datautils.NewStringSet(objectAccounts) //nolint: gochecknoglobals

func records(objectName string) common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		if singleRecordObjects.Has(objectName) {
			record, err := jsonquery.Convertor.ObjectToMap(node)
			if err != nil {
				return nil, err
			}

			return []map[string]any{record}, nil
		}

		records, err := jsonquery.New(node).ArrayRequired(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func nextRecordsURL(baseURL string) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		nextPage, err := jsonquery.New(node, "links", "next_page").StringOptional("href")
		if err != nil {
			return "", err
		}

		if nextPage == nil {
			return "", nil
		}

		res, err := url.JoinPath(baseURL, *nextPage)
		if err != nil {
			return "", err
		}

		return res, nil
	}
}

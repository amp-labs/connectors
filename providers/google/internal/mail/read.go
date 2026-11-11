package mail

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize = "500"

	// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list
	objectNameMessages = "messages"
	// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list
	objectNameDrafts = "drafts"
	// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.threads/list
	objectNameThreads = "threads"

	virtualObjectSentMessages  = "AMPERSAND-sentMessages" // based on "messages"
	virtualObjectInboxMessages = "AMPERSAND-messages"     // based on "messages"
	conditionIsSentLabel       = "in:SENT"
	conditionIsInboxLabel      = "in:INBOX"
)

// nolint:gochecknoglobals
var (
	// objectGroupTypeMessages groups all message-based objects: the core "messages" resource
	// and virtual views like "sentMessages" and "inboxMessages". These virtual objects are
	// the same underlying Gmail messages, just requested with different filters (labels).
	// They share the same schema and fields; only the query differs.
	objectGroupTypeMessages = datautils.NewSet(
		objectNameMessages, virtualObjectSentMessages, virtualObjectInboxMessages,
	)
	paginatedObjects = datautils.NewSet(
		objectNameDrafts, objectNameMessages, objectNameThreads,
		virtualObjectSentMessages, virtualObjectInboxMessages,
	)
	incrementalObjects = datautils.NewSet(
		objectNameDrafts, objectNameMessages, objectNameThreads,
		virtualObjectSentMessages, virtualObjectInboxMessages,
	)
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.getReadUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Add pagination query parameters.
	if paginatedObjects.Has(params.ObjectName) {
		pageSize := readhelper.PageSizeWithDefaultStr(params, defaultPageSize)
		url.WithQueryParam("maxResults", pageSize)

		if params.NextPage != "" {
			url.WithQueryParam("pageToken", params.NextPage.String())
		}
	}

	// nolint:lll
	//
	// Gmail does not expose first-class timestamp filters on list endpoints.
	// Time-based incremental reads must be implemented using the Gmail search DSL
	// via the `q` parameter (e.g. `after:` / `before:`).
	//
	// Although some sources suggest Unix timestamps may work, this behavior is not
	// clearly documented and has been reported as inconsistent:
	// https://stackoverflow.com/questions/56455757/gmail-api-messages-list-q-aftertimestamp-doe-not-work-properly/56482916#56482916
	//
	// The officially documented and UI-supported format uses year/month/day:
	// https://support.google.com/mail/answer/7190
	//
	// The following collection endpoints support the `q` search parameter and
	// therefore can be time-filtered:
	// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list
	// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list
	// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.threads/list
	query := newSearchBoxQuery()
	if incrementalObjects.Has(params.ObjectName) {
		query = query.
			WithSince(params.Since).
			WithUntil(params.Until)
	}

	switch params.ObjectName {
	case virtualObjectSentMessages:
		query = query.WithCondition(conditionIsSentLabel)
	case virtualObjectInboxMessages:
		query = query.WithCondition(conditionIsInboxLabel)
	}

	queryStr := query.String()
	if queryStr != "" {
		url.WithQueryParam("q", queryStr)
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	nativeObjectName := params.ObjectName
	if objectGroupTypeMessages.Has(params.ObjectName) {
		nativeObjectName = objectNameMessages
	}

	responseFieldName := Schemas.LookupArrayFieldName(a.Module(), nativeObjectName)
	identifierLocator := readhelper.NewNestedIdField([]string{responseFieldName}, "id")
	marshaller := readhelper.MakeMarshaledDataFuncWithId(nil, identifierLocator)

	// Gmail list endpoints (messages, drafts, threads) return only a minimal envelope:
	//   - messages: { id, threadId }
	//   - drafts:   { id, message: { id, threadId } }
	//   - threads:  { id, historyId, ... }
	//
	// If the caller requests any field beyond id/threadId on an object that “owns” a full
	// Gmail message (real messages, virtual message views, or drafts), we must fetch the
	// full message bodies and embed them into the list response.
	//
	// Object categories:
	//   1. Core messages:          "messages"
	//   2. Virtual message views:  "AMPERSAND-sentMessages", "AMPERSAND-messages"
	//      (same underlying message object, just filtered by label)
	//   3. Drafts:                 "drafts"
	//      (different top-level object that contains a message)
	//
	// From the caller’s perspective, fields are agnostic to these categories: they just
	// ask for fields on "messages", "sentMessages", "inboxMessages", or "drafts".
	// This function decides when to aggregate full messages into the list response.
	needsFullMessageFields := params.Fields.HasExtra(datautils.NewSet("id", "threadId"))
	isMessageBasedObject := nativeObjectName == objectNameMessages || nativeObjectName == objectNameDrafts

	if needsFullMessageFields && isMessageBasedObject {
		// Fetch full Gmail messages to enrich the minimal list response.
		// For messages and virtual message views, the list items are messages.
		// For drafts, each list item wraps a message in a different structure.
		messages, err := a.fetchMessages(ctx, nativeObjectName, resp)
		if err != nil {
			return nil, err
		}

		if nativeObjectName == objectNameDrafts {
			// Drafts have a different envelope: { id, message: {...} }.
			// We embed the fetched full message into that draft-shaped response.
			marshaller = readhelper.MakeMarshaledSelectedDataFunc(
				draftsEmbedMessageFields(messages),
				draftsEmbedMessageRaw(messages),
			)
		} else {
			// For core messages and virtual message views, the list items are already messages.
			// We just enrich them with the full message payload.
			marshaller = readhelper.MakeMarshaledSelectedDataFunc(
				messagesEmbedMessageFields(messages),
				messagesEmbedMessageRaw(messages),
			)
		}
	}

	return common.ParseResult(resp,
		func(node *ajson.Node) ([]*ajson.Node, error) {
			return jsonquery.New(node).ArrayOptional(responseFieldName)
		},
		makeNextRecordsURL(params),
		marshaller,
		params.Fields,
	)
}

func makeNextRecordsURL(params common.ReadParams) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		if !paginatedObjects.Has(params.ObjectName) {
			// There is no next page.
			return "", nil
		}

		return jsonquery.New(node).StrWithDefault("nextPageToken", "")
	}
}

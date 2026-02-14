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

	objectNameMessages = "messages"
	objectNameDrafts   = "drafts"
)

// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list
// https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.threads/list
var paginatedObjects = datautils.NewSet(objectNameDrafts, objectNameMessages, "threads") // nolint:gochecknoglobals

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.getReadURL(params.ObjectName)
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
	if datautils.NewSet(objectNameDrafts, objectNameMessages, "threads").Has(params.ObjectName) {
		query := newTimeQuery().
			WithSince(params.Since).
			WithUntil(params.Until).
			String()

		if query != "" {
			url.WithQueryParam("q", query)
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	marshaller := common.MakeMarshaledDataFunc(nil)

	// Messages and Drafts objects with extra fields require fetching the full message.
	// See Gmail API: https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list#response-body
	if (params.ObjectName == objectNameMessages || params.ObjectName == objectNameDrafts) &&
		params.Fields.HasExtra(datautils.NewSet("id", "threadId")) {
		messages, err := a.fetchMessages(ctx, params.ObjectName, resp)
		if err != nil {
			return nil, err
		}

		if params.ObjectName == objectNameMessages {
			marshaller = readhelper.MakeMarshaledSelectedDataFunc(
				messagesEmbedMessageFields(params, messages),
				messagesEmbedMessageRaw(messages),
			)
		} else {
			marshaller = readhelper.MakeMarshaledSelectedDataFunc(
				draftsEmbedMessageFields(params, messages),
				draftsEmbedMessageRaw(messages),
			)
		}
	}

	responseFieldName := Schemas.LookupArrayFieldName(a.Module(), params.ObjectName)

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

package mailgun

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// messagesPath is the send-message endpoint. "messages" is a write-only object:
// Mailgun offers no list/read for sent messages, so it has no schema entry.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/messages
const messagesPath = "/v3/{domain_name}/messages"

// listAddressField is the sidecar key in RecordData naming the parent mailing
// list of a lists/members write. It is stripped from the payload before send.
const listAddressField = "list_address"

type writeStyle int

const (
	// writeForm sends the record as application/x-www-form-urlencoded. Mailgun
	// documents most write endpoints as multipart/form-data but accepts
	// URL-encoded forms equivalently (verified live against POST /v3/lists).
	writeForm writeStyle = iota
	// writeQuery sends the record as query parameters with an empty body
	// (the forwards endpoints take no request body).
	writeQuery
)

type writeSpec struct {
	scope objectScope
	style writeStyle
	// canUpdate marks objects with an item-level update endpoint (or, for
	// lists/members, POST upsert semantics). Suppression objects are
	// create/delete only.
	canUpdate bool
	// recordKey is the response envelope key holding the written record
	// ("" means the record fields sit at the response root).
	recordKey string
	// idField is the record field returned as WriteResult.RecordId.
	idField string
}

// objectWriteSpecs is the verified per-object write behaviour, sourced from the
// Mailgun OpenAPI spec (providers/mailgun/metadata/openapi/mailgun.yaml).
// Creates POST the collection path; updates PUT <collection>/<RecordId>
// (except lists/members — see buildMembersRecord).
//
// The JSON variant of the suppression endpoints is their bulk form (a JSON
// array; verified live) and belongs to bulk write, so single-record writes are
// form-encoded.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/intro/
//
//nolint:gochecknoglobals
var objectWriteSpecs = map[string]writeSpec{
	// Domain-scoped. Suppressions are create-only (no update endpoint exists).
	"messages":     {scopeDomain, writeForm, false, "", "id"},
	"bounces":      {scopeDomain, writeForm, false, "", "address"},
	"complaints":   {scopeDomain, writeForm, false, "", "address"},
	"unsubscribes": {scopeDomain, writeForm, false, "", "address"},
	"whitelists":   {scopeDomain, writeForm, false, "", "value"},
	"templates":    {scopeDomain, writeForm, true, "template", "name"},

	// Account-scoped.
	"account/templates": {scopeAccount, writeForm, true, "template", "name"},
	"lists":             {scopeAccount, writeForm, true, "list", "address"},
	"lists/members":     {scopeAccount, writeForm, true, "member", "address"},
	"routes":            {scopeAccount, writeForm, true, "route", "id"},
	"forwards":          {scopeAccount, writeQuery, true, "", "id"},
	"webhooks":          {scopeAccount, writeForm, true, "", "webhook_id"},
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	spec, ok := objectWriteSpecs[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if params.IsUpdate() && !spec.canUpdate {
		// Suppressions and messages have no update endpoint.
		return nil, common.ErrOperationNotSupportedForObject
	}

	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	url, method, err := c.buildWriteURL(params, spec, record)
	if err != nil {
		return nil, err
	}

	return newWriteHTTPRequest(ctx, url, method, spec.style, record)
}

// newWriteHTTPRequest attaches the record in the object's style: a URL-encoded
// form body, or query parameters with an empty body (forwards).
func newWriteHTTPRequest(
	ctx context.Context, url *urlbuilder.URL, method string, style writeStyle, record common.Record,
) (*http.Request, error) {
	var (
		body        []byte
		contentType string
		err         error
	)

	switch style {
	case writeQuery:
		for _, key := range sortedRecordKeys(record) {
			if value := record[key]; value != nil {
				url.WithQueryParam(key, fmt.Sprint(value))
			}
		}
	case writeForm:
		fallthrough
	default:
		body, err = httpkit.EncodeForm(record)
		if err != nil {
			return nil, err
		}

		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

// buildWriteURL resolves the endpoint and method for the write. Creates POST
// the collection path; updates PUT <collection>/<RecordId>. lists/members is
// the exception: its item path would need the parent list as a second
// identifier, which the connector API does not support, so updates reuse POST
// with Mailgun's upsert flag (see buildMembersRecord).
func (c *Connector) buildWriteURL(
	params common.WriteParams, spec writeSpec, record common.Record,
) (*urlbuilder.URL, string, error) {
	path, err := c.writePath(params.ObjectName)
	if err != nil {
		return nil, "", err
	}

	if spec.scope == scopeDomain {
		path, err = c.substituteDomain(path)
		if err != nil {
			return nil, "", err
		}
	}

	if params.ObjectName == "lists/members" {
		path, err = buildMembersRecord(params, record, path)
		if err != nil {
			return nil, "", err
		}

		// Both create and upsert-update POST the collection.
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)

		return url, http.MethodPost, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, "", err
	}

	if params.IsUpdate() {
		url.AddPath(params.RecordId)

		return url, http.MethodPut, nil
	}

	return url, http.MethodPost, nil
}

// buildMembersRecord resolves the lists/members collection path from the
// required list_address sidecar in RecordData (stripped from the payload) and,
// on update, folds RecordId into the payload with Mailgun's upsert flag.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/mailing-lists
func buildMembersRecord(params common.WriteParams, record common.Record, path string) (string, error) {
	listAddress, ok := record[listAddressField].(string)
	if !ok || listAddress == "" {
		return "", fmt.Errorf("%w: lists/members requires a %q field naming the parent mailing list",
			common.ErrMissingRecordData, listAddressField)
	}

	delete(record, listAddressField)

	if params.IsUpdate() {
		record["address"] = params.RecordId
		record["upsert"] = "yes"
	}

	return strings.ReplaceAll(path, listAddressPlaceholder, listAddress), nil
}

func sortedRecordKeys(record common.Record) []string {
	keys := make([]string, 0, len(record))
	for key := range record {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// writePath maps the object to its write collection path. All curated write
// objects share their read path; messages is write-only and hardcoded.
func (c *Connector) writePath(objectName string) (string, error) {
	if objectName == "messages" {
		return messagesPath, nil
	}

	return metadata.Schemas.LookupURLPath(c.ProviderContext.Module(), objectName)
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	spec := objectWriteSpecs[params.ObjectName]

	node := body

	if spec.recordKey != "" {
		envelope, err := jsonquery.New(body).ObjectOptional(spec.recordKey)
		if err != nil {
			return nil, err
		}

		if envelope != nil {
			node = envelope
		}
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	recordID := params.RecordId
	if value, ok := data[spec.idField]; ok && value != nil {
		recordID = fmt.Sprint(value)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

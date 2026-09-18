package mailgun

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// collectionPathOverrides names the write/delete collection for objects whose
// mutation endpoint differs from (or has no) read path. Every other object
// shares its read path, looked up from the static schema.
//
//   - messages is write-only: Mailgun offers no list/read for sent messages, so
//     it has no schema entry.
//     https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/messages
//   - alerts/settings is read from the /v1/alerts/settings envelope ("events"
//     key) but created, updated and deleted at /v1/alerts/settings/events.
//     https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/alerts
//
//nolint:gochecknoglobals
var collectionPathOverrides = map[string]string{
	"messages":        "/v3/{domain_name}/messages",
	"alerts/settings": "/v1/alerts/settings/events",
}

// listAddressField is the sidecar key in RecordData naming the parent mailing
// list of a lists/members write. It is stripped from the payload before send.
const listAddressField = "list_address"

type writeStyle int

const (
	// writeForm sends the record as application/x-www-form-urlencoded. Mailgun
	// documents most write endpoints as multipart/form-data but accepts
	// URL-encoded forms equivalently (verified live against POST /v3/lists).
	writeForm writeStyle = iota
	// writeQuery sends the record as query parameters with an empty body (the
	// forwards endpoints take no request body). Slices repeat the parameter,
	// as the spec requires for forward.url / forward.recipient.
	writeQuery
	// writeJSON sends the record as an application/json body (the alerts
	// settings endpoints are JSON-only).
	writeJSON
)

type writeSpec struct {
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
// (except lists/members — see buildMembersRecord). Records are accepted in the
// shape Read returns; the two objects whose write fields are named differently
// from their read fields are reshaped first (see adaptWriteRecord).
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
	"messages":     {writeForm, false, "", "id"},
	"bounces":      {writeForm, false, "", "address"},
	"complaints":   {writeForm, false, "", "address"},
	"unsubscribes": {writeForm, false, "", "address"},
	"whitelists":   {writeForm, false, "", "value"},
	"templates":    {writeForm, true, "template", "name"},

	// Account-scoped.
	"account/templates": {writeForm, true, "template", "name"},
	"lists":             {writeForm, true, "list", "address"},
	"lists/members":     {writeForm, true, "member", "address"},
	"routes":            {writeForm, true, "route", "id"},
	"forwards":          {writeQuery, true, "", "id"},
	"webhooks":          {writeForm, true, "", "webhook_id"},
	"alerts/settings":   {writeJSON, true, "", "id"},
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

	// Work on a copy: GetRecord hands back the caller's own map when RecordData
	// is already a map, and the reshaping below renames and removes keys.
	record = adaptWriteRecord(params.ObjectName, maps.Clone(record))

	url, method, err := c.buildWriteURL(params, record)
	if err != nil {
		return nil, err
	}

	return newWriteHTTPRequest(ctx, url, method, spec.style, record)
}

// adaptWriteRecord reshapes a record given in the shape Read returns into the
// shape the write endpoint accepts, so a ReadResult can feed a Write. Only two
// objects differ; every other object's write fields match its read fields.
//
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/routes
// API reference: https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/forwards
func adaptWriteRecord(objectName string, record common.Record) common.Record {
	switch objectName {
	case "routes":
		// Read returns the executed actions as "actions": [...]; POST and PUT
		// take the same list as a repeated "action" form field.
		renameKey(record, "actions", "action")
	case "forwards":
		expandForward(record)
	}

	return record
}

// renameKey moves record[from] to record[to] unless the caller already supplied
// the destination key, in which case the read-shaped key is dropped.
func renameKey(record common.Record, from, to string) {
	value, ok := record[from]
	if !ok {
		return
	}

	delete(record, from)

	if _, exists := record[to]; !exists {
		record[to] = value
	}
}

// forwardKeys maps the keys of the nested read-side "forward" object to the
// dotted, repeatable query parameters the forwards write endpoints accept.
//
//nolint:gochecknoglobals
var forwardKeys = map[string]string{
	"urls":       "forward.url",
	"recipients": "forward.recipient",
	"store":      "forward.store",
}

// expandForward flattens a read-shaped {"forward": {"urls": [...],
// "recipients": [...], "store": "..."}} into forward.url, forward.recipient and
// forward.store; the query encoder then repeats each list element.
func expandForward(record common.Record) {
	nested, ok := record["forward"].(map[string]any)
	if !ok {
		return
	}

	delete(record, "forward")

	for key, param := range forwardKeys {
		value, found := nested[key]
		if !found {
			continue
		}

		if _, exists := record[param]; !exists {
			record[param] = value
		}
	}
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
		if err = applyQueryRecord(url, record); err != nil {
			return nil, err
		}
	case writeJSON:
		body, err = json.Marshal(record)
		if err != nil {
			return nil, err
		}

		contentType = "application/json"
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
	params common.WriteParams, record common.Record,
) (*urlbuilder.URL, string, error) {
	path, err := c.collectionPath(params.ObjectName)
	if err != nil {
		return nil, "", err
	}

	path, err = c.substituteDomain(path)
	if err != nil {
		return nil, "", err
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

// applyQueryRecord carries the record as query parameters using the same
// encoding as the form body (slices repeat the key, nested objects are JSON),
// so repeatable parameters such as forward.url arrive as repeats rather than a
// stringified slice.
func applyQueryRecord(url *urlbuilder.URL, record common.Record) error {
	encoded, err := httpkit.EncodeForm(record)
	if err != nil {
		return err
	}

	values, err := neturl.ParseQuery(string(encoded))
	if err != nil {
		return err
	}

	for key, list := range values {
		url.WithQueryParamList(key, list)
	}

	return nil
}

// collectionPath maps the object to the collection its records are created in
// and deleted from (item paths append RecordId). Objects default to their read
// path; see collectionPathOverrides for the exceptions.
func (c *Connector) collectionPath(objectName string) (string, error) {
	if path, ok := collectionPathOverrides[objectName]; ok {
		return path, nil
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

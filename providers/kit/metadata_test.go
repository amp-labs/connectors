// nolint
package kit

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	broadcastsresponse := testutils.DataFromFile(t, "broadcasts.json")
	customfieldsresponse := testutils.DataFromFile(t, "custom_fields.json")
	emailtemplatesresponse := testutils.DataFromFile(t, "email_templates.json")
	formsresponse := testutils.DataFromFile(t, "forms.json")
	purchasesresponse := testutils.DataFromFile(t, "purchases.json")
	sequencesresponse := testutils.DataFromFile(t, "sequences.json")
	segmentsresponse := testutils.DataFromFile(t, "segments.json")
	subscribersresponse := testutils.DataFromFile(t, "subscribers.json")
	tagsresponse := testutils.DataFromFile(t, "tags.json")
	webhooksresponse := testutils.DataFromFile(t, "webhooks.json")

	tests := []testroutines.Metadata{
		{
			Name:         "Object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple object with metadata",
			Input: []string{"broadcasts", "custom_fields", "forms", "subscribers", "tags", "email_templates", "purchases", "segments", "sequences", "webhooks"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/v4/broadcasts"),
					Then: mockserver.Response(http.StatusOK, broadcastsresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/custom_fields"),
					Then: mockserver.Response(http.StatusOK, customfieldsresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/email_templates"),
					Then: mockserver.Response(http.StatusOK, emailtemplatesresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/forms"),
					Then: mockserver.Response(http.StatusOK, formsresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/purchases"),
					Then: mockserver.Response(http.StatusOK, purchasesresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/sequences"),
					Then: mockserver.Response(http.StatusOK, sequencesresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/subscribers"),
					Then: mockserver.Response(http.StatusOK, subscribersresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/segments"),
					Then: mockserver.Response(http.StatusOK, segmentsresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/tags"),
					Then: mockserver.Response(http.StatusOK, tagsresponse),
				}, {
					If:   mockcond.PathSuffix("/v4/webhooks"),
					Then: mockserver.Response(http.StatusOK, webhooksresponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"broadcasts": {
						DisplayName: "Broadcasts",
						FieldsMap: map[string]string{
							"content":           "content",
							"created_at":        "created_at",
							"description":       "description",
							"email_address":     "email_address",
							"email_template":    "email_template",
							"id":                "id",
							"preview_text":      "preview_text",
							"public":            "public",
							"published_at":      "published_at",
							"send_at":           "send_at",
							"subject":           "subject",
							"subscriber_filter": "subscriber_filter",
							"thumbnail_alt":     "thumbnail_alt",
							"thumbnail_url":     "thumbnail_url",
						},
					},
					"custom_fields": {
						DisplayName: "Custom Fields",
						FieldsMap: map[string]string{
							"id":    "id",
							"key":   "key",
							"label": "label",
							"name":  "name",
						},
					},
					"email_templates": {
						DisplayName: "Email Templates",
						FieldsMap: map[string]string{
							"category":   "category",
							"id":         "id",
							"is_default": "is_default",
							"name":       "name",
						},
					},
					"forms": {
						DisplayName: "Forms",
						FieldsMap: map[string]string{
							"archived":   "archived",
							"created_at": "created_at",
							"embed_js":   "embed_js",
							"embed_url":  "embed_url",
							"format":     "format",
							"id":         "id",
							"name":       "name",
							"type":       "type",
							"uid":        "uid",
						},
					},
					"purchases": {
						DisplayName: "Purchases",
						FieldsMap: map[string]string{
							"currency":         "currency",
							"discount":         "discount",
							"email_address":    "email_address",
							"id":               "id",
							"products":         "products",
							"status":           "status",
							"subtotal":         "subtotal",
							"tax":              "tax",
							"total":            "total",
							"transaction_id":   "transaction_id",
							"transaction_time": "transaction_time",
						},
					},
					"segments": {
						DisplayName: "Segments",
						FieldsMap: map[string]string{
							"id":         "id",
							"name":       "name",
							"created_at": "created_at",
						},
					},
					"sequences": {
						DisplayName: "Sequences",
						FieldsMap: map[string]string{
							"created_at": "created_at",
							"hold":       "hold",
							"id":         "id",
							"name":       "name",
							"repeat":     "repeat",
						},
					},
					"subscribers": {
						DisplayName: "Subscribers",
						FieldsMap: map[string]string{
							"created_at":    "created_at",
							"email_address": "email_address",
							"fields":        "fields",
							"first_name":    "first_name",
							"id":            "id",
							"state":         "state",
						},
					},
					"tags": {
						DisplayName: "Tags",
						FieldsMap: map[string]string{
							"created_at": "created_at",
							"id":         "id",
							"name":       "name",
						},
					},
					"webhooks": {
						DisplayName: "Webhooks",
						FieldsMap: map[string]string{
							"account_id": "account_id",
							"event":      "event",
							"id":         "id",
							"target_url": "target_url",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine.
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)

	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(serverURL)

	return connector, nil
}

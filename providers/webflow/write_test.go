package webflow

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseRedirectCreate := testutils.DataFromFile(t, "write/redirect-create.json")
	responseCollectionUpdate := testutils.DataFromFile(t, "write/collection-update.json")
	responsePageUpdate := testutils.DataFromFile(t, "write/page-update.json")
	responseOrderUpdate := testutils.DataFromFile(t, "write/order-update.json")
	responseProductCreate := testutils.DataFromFile(t, "write/product-create.json")
	responseProductUpdate := testutils.DataFromFile(t, "write/product-update.json")
	responseCustomFontUpdate := testutils.DataFromFile(t, "write/custom-font-update.json")
	responseWebhookCreate := testutils.DataFromFile(t, "write/webhook-create.json")
	responseScriptCreate := testutils.DataFromFile(t, "write/script-create.json")
	responseValidation := testutils.DataFromFile(t, "write/error-validation.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectRedirects},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.WriteParams{ObjectName: "items", RecordData: map[string]any{"name": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Read-only object is not writable",
			Input:        common.WriteParams{ObjectName: objectForms, RecordData: map[string]any{"name": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create-only object rejects update",
			Input: common.WriteParams{
				ObjectName: objectWebhooks, RecordId: "webhook-1",
				RecordData: map[string]any{"url": "https://example.com"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Update-only object rejects create",
			Input:        common.WriteParams{ObjectName: objectPages, RecordData: map[string]any{"title": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create redirect posts to the site collection",
			Input: common.WriteParams{
				ObjectName: objectRedirects,
				RecordData: map[string]any{"fromUrl": "/old-path", "toUrl": "/new-path"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testSitePath + "/redirects"),
					mockcond.Body(`{"fromUrl":"/old-path","toUrl":"/new-path"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseRedirectCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "redirect-1",
				Data:     map[string]any{"id": "redirect-1", "fromUrl": "/old-path", "toUrl": "/new-path"},
			},
		},
		{
			Name: "Update collection patches the item path outside the site",
			Input: common.WriteParams{
				ObjectName: objectCollections,
				RecordId:   "collection-1",
				RecordData: map[string]any{"displayName": "Blog Posts (updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v2/collections/collection-1"),
					mockcond.Body(`{"displayName":"Blog Posts (updated)"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCollectionUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "collection-1",
				Data:     map[string]any{"id": "collection-1", "displayName": "Blog Posts (updated)"},
			},
		},
		{
			Name: "Update page uses PUT",
			Input: common.WriteParams{
				ObjectName: objectPages,
				RecordId:   "page-1",
				RecordData: map[string]any{"title": "Home (updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v2/pages/page-1"),
				},
				Then: mockserver.Response(http.StatusOK, responsePageUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "page-1",
				Data:     map[string]any{"id": "page-1", "title": "Home (updated)"},
			},
		},
		{
			Name: "Update order returns orderId",
			Input: common.WriteParams{
				ObjectName: objectOrders,
				RecordId:   "a1b2-c3d4",
				RecordData: map[string]any{"comment": "Leave at the door"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path(testSitePath + "/orders/a1b2-c3d4"),
				},
				Then: mockserver.Response(http.StatusOK, responseOrderUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "a1b2-c3d4",
				Data:     map[string]any{"orderId": "a1b2-c3d4", "comment": "Leave at the door"},
			},
		},
		{
			Name: "Create product in API shape passes through",
			Input: common.WriteParams{
				ObjectName: objectProducts,
				RecordData: map[string]any{
					"product": map[string]any{"fieldData": map[string]any{"name": "Five-Star Hat", "slug": "five-star-hat"}},
					"sku":     map[string]any{"fieldData": map[string]any{"name": "Five-Star Hat", "price": map[string]any{"value": 2500, "unit": "USD"}}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testSitePath + "/products"),
					mockcond.Body(`{"product":{"fieldData":{"name":"Five-Star Hat","slug":"five-star-hat"}},` +
						`"sku":{"fieldData":{"name":"Five-Star Hat","price":{"unit":"USD","value":2500}}}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseProductCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "product-1",
				Data: map[string]any{
					"product": map[string]any{
						"id":          "product-1",
						"isDraft":     false,
						"isArchived":  false,
						"lastUpdated": "2026-10-06T09:00:00.000Z",
						"fieldData":   map[string]any{"name": "Five-Star Hat", "slug": "five-star-hat"},
					},
				},
			},
		},
		{
			Name: "Create product from the flattened read shape wraps product and first sku",
			Input: common.WriteParams{
				ObjectName: objectProducts,
				RecordData: map[string]any{
					"fieldData":     map[string]any{"name": "Five-Star Hat", "slug": "five-star-hat"},
					"isDraft":       false,
					"publishStatus": "staging",
					"skus": []any{
						map[string]any{"fieldData": map[string]any{"name": "Five-Star Hat", "price": map[string]any{"value": 2500, "unit": "USD"}}},
						map[string]any{"fieldData": map[string]any{"name": "Five-Star Hat XL"}},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testSitePath + "/products"),
					mockcond.Body(`{"product":{"fieldData":{"name":"Five-Star Hat","slug":"five-star-hat"},"isDraft":false},` +
						`"publishStatus":"staging",` +
						`"sku":{"fieldData":{"name":"Five-Star Hat","price":{"unit":"USD","value":2500}}}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseProductCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "product-1",
				Data: map[string]any{
					"product": map[string]any{
						"id":          "product-1",
						"isDraft":     false,
						"isArchived":  false,
						"lastUpdated": "2026-10-06T09:00:00.000Z",
						"fieldData":   map[string]any{"name": "Five-Star Hat", "slug": "five-star-hat"},
					},
				},
			},
		},
		{
			Name: "Update product from flattened fields without skus",
			Input: common.WriteParams{
				ObjectName: objectProducts,
				RecordId:   "product-1",
				RecordData: map[string]any{
					"fieldData": map[string]any{"name": "Five-Star Hat (updated)"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path(testSitePath + "/products/product-1"),
					mockcond.Body(`{"product":{"fieldData":{"name":"Five-Star Hat (updated)"}}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseProductUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "product-1",
				Data:     map[string]any{"id": "product-1", "fieldData": map[string]any{"name": "Five-Star Hat (updated)", "slug": "five-star-hat"}},
			},
		},
		{
			Name: "Update custom font reads the id from the customFont envelope",
			Input: common.WriteParams{
				ObjectName: objectCustomFonts,
				RecordId:   "font-1",
				RecordData: map[string]any{"fontDisplay": "swap"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path(testSitePath + "/custom_fonts/font-1"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomFontUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "font-1",
				Data: map[string]any{
					"customFont": map[string]any{
						"id": "font-1", "fontFamily": "Inter", "weight": "400", "italic": false, "fontDisplay": "swap",
					},
				},
			},
		},
		{
			Name: "Create webhook",
			Input: common.WriteParams{
				ObjectName: objectWebhooks,
				RecordData: map[string]any{"triggerType": "form_submission", "url": "https://example.com/hooks/webflow"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testSitePath + "/webhooks"),
				},
				Then: mockserver.Response(http.StatusCreated, responseWebhookCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "webhook-1",
				Data:     map[string]any{"id": "webhook-1", "triggerType": "form_submission"},
			},
		},
		{
			Name: "Registered script with sourceCode goes to the inline endpoint",
			Input: common.WriteParams{
				ObjectName: objectRegisteredScripts,
				RecordData: map[string]any{
					"displayName": "CMS Slider", "version": "1.0.0", "sourceCode": "alert('hi')",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testSitePath + "/registered_scripts/inline"),
				},
				Then: mockserver.Response(http.StatusCreated, responseScriptCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cms_slider",
				Data:     map[string]any{"id": "cms_slider", "displayName": "CMS Slider"},
			},
		},
		{
			Name: "Registered script without sourceCode goes to the hosted endpoint",
			Input: common.WriteParams{
				ObjectName: objectRegisteredScripts,
				RecordData: map[string]any{
					"displayName": "CMS Slider", "version": "1.0.0",
					"hostedLocation": "https://cdn.example.com/slider.js", "integrityHash": "sha384-abc",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(testSitePath + "/registered_scripts/hosted"),
				},
				Then: mockserver.Response(http.StatusCreated, responseScriptCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cms_slider",
				Data:     map[string]any{"id": "cms_slider", "displayName": "CMS Slider"},
			},
		},
		{
			Name: "Validation error surfaces as caller error",
			Input: common.WriteParams{
				ObjectName: objectWebhooks,
				RecordData: map[string]any{},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseValidation),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

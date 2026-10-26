package reply

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseContactCreate := testutils.DataFromFile(t, "write/contact-create.json")
	responseContactUpdate := testutils.DataFromFile(t, "write/contact-update.json")
	responseCustomFieldUpdate := testutils.DataFromFile(t, "write/custom-field-update.json")
	responseAccountCreate := testutils.DataFromFile(t, "write/account-create.json")
	responseAccountUpdate := testutils.DataFromFile(t, "write/account-update.json")
	responseCustomFields := testutils.DataFromFile(t, "read/custom-fields/definitions.json")
	responseCollidingCustomFields := testutils.DataFromFile(t, "read/custom-fields/definitions-colliding.json")
	responseAccountCustomFields := testutils.DataFromFile(t, "read/custom-fields/definitions-accounts.json")
	responseSequenceFolderCreate := testutils.DataFromFile(t, "write/sequence-folder-create.json")
	errorMissingEmail := testutils.DataFromFile(t, "write/err-missing-email.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Read-only object is rejected by the registry",
			Input:        common.WriteParams{ObjectName: "inbox/threads", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Create contact via POST",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contacts"),
				},
				Then: mockserver.Response(http.StatusCreated, responseContactCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Errors:   nil,
				Data: map[string]any{
					"email":     "probe.write@example.com",
					"firstName": "Probe",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact via PATCH",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "761575530",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/761575530"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Errors:   nil,
				Data: map[string]any{
					"title": "Updated",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update custom field via PUT",
			Input: common.WriteParams{
				ObjectName: "custom-fields",
				RecordId:   "150697",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3/custom-fields/150697"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomFieldUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "150697",
				Errors:   nil,
				Data: map[string]any{
					"title": "Deal Stage",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contact customFields from a ReadResult round-trip: key becomes id",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"customFields": []any{map[string]any{
						"key":   "150697",
						"value": "probe",
					}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/759458702"),
					mockcond.Body(`{"customFields":[{"id":150697,"value":"probe"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data: map[string]any{
					"email": "probe.write@example.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contact custom field named by title is sent as a customFields entry",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"firstName": "Jane",
					// Deal Stage is a custom field, named by its title.
					"Deal Stage": "won",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/v3/custom-fields"),
						mockcond.QueryParam("entityType", "contact"),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/v3/contacts/759458702"),
						mockcond.Body(`{"firstName":"Jane","customFields":[{"id":150697,"value":"won"}]}`),
					},
					Then: mockserver.Response(http.StatusOK, responseContactUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Lowercased ReadResult field keys still match the custom field title",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"firstname": "Jane",
					// Deal Stage is a custom field, lowercased as ReadResult.Fields carries it.
					"deal stage": "lost",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/v3/contacts/759458702"),
						mockcond.Body(`{"firstname":"Jane","customFields":[{"id":150697,"value":"lost"}]}`),
					},
					Then: mockserver.Response(http.StatusOK, responseContactUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Title-named value wins over the customFields entry of the same field",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"customfields": []any{map[string]any{"key": "150697", "value": "old"}},
					// Deal Stage is a custom field; it wins over the customFields entry below.
					"deal stage": "new",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/v3/contacts/759458702"),
						mockcond.Body(`{"customFields":[{"id":150697,"value":"new"}]}`),
					},
					Then: mockserver.Response(http.StatusOK, responseContactUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create contact sends custom fields by key, the only form Reply keeps on create",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName": "Jane",
					"customFields": []any{
						map[string]any{"id": float64(150697), "value": "by-id"},
						map[string]any{"name": "Other Field", "value": "by-name"},
						map[string]any{"key": "150699", "value": "by-key"},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contacts"),
					mockcond.Body(`{"firstName":"Jane","customFields":[` +
						`{"key":"150697","value":"by-id"},` +
						`{"key":"Other Field","value":"by-name"},` +
						`{"key":"150699","value":"by-key"}]}`),
				},
				Then: mockserver.Response(http.StatusCreated, responseContactCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact references custom fields by id or name, never by key",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"customFields": []any{
						map[string]any{"key": "150697", "value": "from-key-id"},
						map[string]any{"key": "Other Field", "value": "from-key-title"},
						map[string]any{"name": "Third Field", "value": "by-name"},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/759458702"),
					mockcond.Body(`{"customFields":[` +
						`{"id":150697,"value":"from-key-id"},` +
						`{"name":"Other Field","value":"from-key-title"},` +
						`{"name":"Third Field","value":"by-name"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create contact from a title-named value uses the key form",
			Input: common.WriteParams{
				ObjectName: "contacts",
				// Deal Stage is a custom field, named by its title.
				RecordData: map[string]any{"firstName": "Jane", "Deal Stage": "won"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v3/contacts"),
						mockcond.Body(`{"firstName":"Jane","customFields":[{"key":"150697","value":"won"}]}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseContactCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Built-in fields and write-only properties never trigger a definitions lookup",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName": "Jane",
					"industry":  "Software",
					"phone2":    "+1234567890",
				},
			},
			// No custom-fields case: a lookup would hit the 500 default and fail the write.
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v3/contacts"),
						mockcond.Body(`{"firstName":"Jane","industry":"Software","phone2":"+1234567890"}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseContactCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Property matching no custom field passes through untouched",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName": "Jane",
					"nickname":  "JD",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v3/contacts"),
						mockcond.Body(`{"firstName":"Jane","nickname":"JD"}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseContactCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Custom field titled like a built-in field is not rewritten",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordData: map[string]any{
					"firstName": "Jane",
					"Industry":  "Software",
					// Deal Stage is a custom field, named by its title.
					"Deal Stage": "won",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCollidingCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v3/contacts"),
						mockcond.Body(`{"firstName":"Jane","Industry":"Software",` +
							`"customFields":[{"key":"150697","value":"won"}]}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseContactCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create account sends account custom fields by id with string values",
			Input: common.WriteParams{
				ObjectName: "contact-accounts",
				RecordData: map[string]any{
					"name": "Globex",
					// Renewal Date and Seats are custom account fields, named by their titles.
					"Renewal Date": "2027-06-30",
					"Seats":        float64(40),
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/v3/custom-fields"),
						mockcond.QueryParam("entityType", "account"),
					},
					Then: mockserver.Response(http.StatusOK, responseAccountCustomFields),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v3/contact-accounts"),
						mockcond.Body(`{"name":"Globex","customFields":[` +
							`{"id":154800,"value":"2027-06-30"},{"id":154801,"value":"40"}]}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseAccountCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "88002",
				Data:     map[string]any{"name": "Globex"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Account customFields from a ReadResult round-trip: key becomes id",
			Input: common.WriteParams{
				ObjectName: "contact-accounts",
				RecordId:   "88001",
				RecordData: map[string]any{
					"customfields": []any{map[string]any{
						"key":   "154800",
						"value": "2028-01-01",
					}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v3/contact-accounts/88001"),
					mockcond.Body(`{"customFields":[{"id":154800,"value":"2028-01-01"}]}`),
				},
				Then: mockserver.Response(http.StatusOK, responseAccountUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "88001",
				Data:     map[string]any{"name": "Acme"},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Custom field definitions failure fails the write instead of dropping the value",
			Input: common.WriteParams{
				ObjectName: "contacts",
				// Deal Stage is a custom field, named by its title.
				RecordData: map[string]any{"firstName": "Jane", "Deal Stage": "won"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/custom-fields"),
				Then:  mockserver.Response(http.StatusInternalServerError, nil),
			}.Server(),
			ExpectedErrs: []error{common.ErrResolvingCustomFields, common.ErrServer},
		},
		{
			Name: "Entry without a value is sent without one instead of clearing the field",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: map[string]any{
					"customFields": []any{map[string]any{
						"key":              "150697",
						"enrichmentStatus": "succeeded",
					}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/759458702"),
					mockcond.Body(`{"customFields":[{"id":150697}]}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactUpdate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "761575530",
				Data:     map[string]any{"email": "probe.write@example.com"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create sequence folder returns a GUID record id",
			Input: common.WriteParams{ObjectName: "sequence-folders", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/sequence-folders"),
				},
				Then: mockserver.Response(http.StatusCreated, responseSequenceFolderCreate),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "42b54f7a-b054-480b-ad1b-6f924c05ea4b",
				Errors:   nil,
				Data: map[string]any{
					"name": "Fixture Capture Folder",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Validation problem on create carries pointer detail",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorMissingEmail),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Validation failed: /firstName FirstName is required."),
			},
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

package associations

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestBatchCreate(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseContactsSchema := testutils.DataFromFile(t, "contacts-schema-resp.json")
	responseAssociationCreated := testutils.DataFromFile(t, "contact-to-company-resp.json")
	errorAssociationCreation := testutils.DataFromFile(t, "err-contact-to-contact.json")

	var (
		// Relationships correspond to the numbers used by Hubspot.
		// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/guide#deal-to-object
		relContactToDeal    = 4
		relContactToCompany = 279
	)

	paramsNormal := NewBatchCreateParams("contacts").
		WithAssociation("contact1", &BatchInput{
			{To: Identifier{ID: "deal_id"}, Types: []Type{{ID: relContactToDeal}}},
		})

	tests := []batchCreateTestCase{
		{
			Name:  "Schema fetch failure returns a warning",
			Input: paramsNormal,
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/crm-object-schemas/2026-03/schemas/contacts"),
				},
				Then: mockserver.ResponseString(http.StatusNotFound, `test-message`),
			}.Server(),
			Comparator: batchCreateComparator,
			Expected: &BatchCreateResult{
				Success: false,
				Errors: []error{testutils.StringError(
					"batch create associations failure: HTTP status 404: retryable error: test-message")},
			},
			ExpectedErrs: nil, // no errors
		},
		{
			Name: "Warning for unresolved association type ID",
			Input: NewBatchCreateParams("contacts").
				WithAssociation("contact_1", &BatchInput{
					{To: Identifier{ID: "deal_id"}, Types: []Type{{ID: relContactToDeal}}},
					{To: Identifier{ID: "unknown_identifier"}, Types: []Type{{ID: 999}}}, // invalid rel id
				}),
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/crm-object-schemas/2026-03/schemas/contacts"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/associations/2026-03/0-1/0-3/batch/create"),
					},
					Then: mockserver.Response(http.StatusOK, []byte(`{"results":[]}`)), // dummy success response
				}},
			}.Server(),
			Comparator: batchCreateComparator,
			Expected: &BatchCreateResult{
				Success: false,
				Errors: []error{
					fmt.Errorf("batch create associations failure: %w",
						fmt.Errorf("%w: objectName(contacts) identifiers(999)", ErrUnresolvedAssociationTypeID)),
				},
			},
		},
		{
			Name:  "Partial failure with 207 Multi-Status",
			Input: paramsNormal,
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/crm-object-schemas/2026-03/schemas/contacts"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/associations/2026-03/0-1/0-3/batch/create"),
					},
					Then: mockserver.Response(http.StatusMultiStatus, errorAssociationCreation),
				}},
			}.Server(),
			Comparator: batchCreateComparator,
			Expected: &BatchCreateResult{
				Success: false,
				Errors: []error{
					fmt.Errorf("batch create associations failure: %w", BatchCreateResponseError{
						Category:   "VALIDATION_ERROR",
						Message:    "unknown_id_123456798 is not a valid ID",
						fromObject: "0-1",
						toObject:   "0-3",
					}),
					fmt.Errorf("batch create associations failure: %w", BatchCreateResponseError{
						Category:   "VALIDATION_ERROR",
						Message:    "Object 211762804612 cannot be associated to itself (211762804612)",
						fromObject: "0-1",
						toObject:   "0-3",
					}),
				},
			},
		},
		{
			Name: "Successful creation of multiple associations",
			Input: NewBatchCreateParams("contacts").
				WithAssociation("contact_id", &BatchInput{
					{To: Identifier{ID: "deal_id"}, Types: []Type{{ID: relContactToDeal}}},
					{To: Identifier{ID: "company_id"}, Types: []Type{{ID: relContactToCompany}}},
				}),
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/crm-object-schemas/2026-03/schemas/contacts"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsSchema),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/associations/2026-03/0-1/0-3/batch/create"),
						mockcond.Body(`{"inputs":[{
							"from":{"id":"contact_id"},
							"to":{"id":"deal_id"},
							"types":[{"associationTypeId":4,"associationCategory":""}]
						}]}`),
					},
					Then: mockserver.Response(http.StatusOK, responseAssociationCreated),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/crm/associations/2026-03/0-1/0-2/batch/create"),
						mockcond.Body(`{"inputs":[{
							"from":{"id":"contact_id"},
							"to":{"id":"company_id"},
							"types":[{"associationTypeId":279,"associationCategory":""}]
						}]}`),
					},
					Then: mockserver.Response(http.StatusOK, responseAssociationCreated),
				}},
			}.Server(),
			Comparator: batchCreateComparator,
			Expected: &BatchCreateResult{
				Success: true,
				Errors:  nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (*Strategy, error) {
				return constructTestStrategy(tt.Server.URL)
			})
		})
	}
}

func constructTestStrategy(serverURL string) (*Strategy, error) {
	transport, err := components.NewTransport(providers.Hubspot, common.ConnectorParams{
		Module:              providers.ModuleHubspotCRM,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestMockServerBaseURL(serverURL)

	return NewStrategy(transport.JSONHTTPClient(), transport.ModuleInfo(), transport.ProviderInfo()), nil
}

type (
	testCaseTypeBatchCreate = testroutines.TestCase[*BatchCreateParams, *BatchCreateResult]
	batchCreateTestCase     testCaseTypeBatchCreate
)

func (c batchCreateTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Strategy]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.BatchCreate(t.Context(), c.Input)
	testCaseTypeBatchCreate(c).Validate(t, err, output)
}

func batchCreateComparator(_ string, actual, expected *BatchCreateResult) bool {
	different := actual.Success != expected.Success ||
		!mockutils.ErrorNormalizedComparator.EachErrorEquals(
			goutils.ToAnySlice(expected.Errors),
			goutils.ToAnySlice(actual.Errors),
		)

	return !different
}

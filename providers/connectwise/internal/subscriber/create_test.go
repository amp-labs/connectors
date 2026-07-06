package subscriber

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestCreate(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	requestWebhookForContacts := testutils.DataFromFile(t, "create/contacts/request.json")
	responseWebhookForContacts := testutils.DataFromFile(t, "create/contacts/response.json")
	requestWebhookForTickets := testutils.DataFromFile(t, "create/tickets/request.json")
	responseWebhookForTickets := testutils.DataFromFile(t, "create/tickets/response.json")
	errorCreateBadRequest1 := testutils.DataFromFile(t, "create/create-webhook-bad-request-1.json")
	errorCreateBadRequest2 := testutils.DataFromFile(t, "create/create-webhook-bad-request-2.json")
	deleteWebhookFailed := testutils.DataFromFile(t, "create/remove-webhook-failed.json")

	eventTypesCUD := []common.SubscriptionEventType{
		common.SubscriptionEventTypeCreate,
		common.SubscriptionEventTypeUpdate,
		common.SubscriptionEventTypeDelete,
	}

	tests := []testconn.TestCaseSubscribe{
		{
			Name: "Creating subscription to contacts and tickets successfully",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts":        {Events: eventTypesCUD},
					"project/tickets": {Events: eventTypesCUD},
				},
				Request: &Request{
					WebhookURL: "https://test.com/webhook",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForContacts),
					},
					Then: mockserver.Response(http.StatusOK, responseWebhookForContacts),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForTickets),
					},
					Then: mockserver.Response(http.StatusOK, responseWebhookForTickets),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareResult),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						"project/tickets": {
							ID:         26552,
							WebhookURL: "https://test.com/webhook?recordId=",
							ObjectType: "ticket",
						},
						"contacts": {
							ID:         26559,
							WebhookURL: "https://test.com/webhook?recordId=",
							ObjectType: "contact",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts": {
						Events:            eventTypesCUD,
						WatchFields:       nil,
						WatchFieldsAll:    false,
						PassThroughEvents: nil,
					},
					"project/tickets": {
						Events:            eventTypesCUD,
						WatchFields:       nil,
						WatchFieldsAll:    false,
						PassThroughEvents: nil,
					},
				},
				Status: common.SubscriptionStatusSuccess,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Ticket creation failed and then failed to rollback contacts.",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts":        {Events: eventTypesCUD},
					"project/tickets": {Events: eventTypesCUD},
				},
				Request: &Request{
					WebhookURL: "https://test.com/webhook",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForContacts),
					},
					Then: mockserver.Response(http.StatusOK, responseWebhookForContacts),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForTickets),
					},
					Then: mockserver.Response(http.StatusBadRequest, errorCreateBadRequest1),
				}, {
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks/26559"),
					},
					Then: mockserver.Response(http.StatusBadRequest, deleteWebhookFailed),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareResult),
			Expected: &common.SubscriptionResult{
				Result: &Result{
					ObjectWebhooks: map[common.ObjectName]SubscriptionResource{
						// Contacts webhook still exists.
						"contacts": {
							ID:         26559,
							WebhookURL: "https://test.com/webhook?recordId=",
							ObjectType: "contact",
						},
					},
				},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts": {
						Events:            eventTypesCUD, // contact subscription remains because couldn't remove it.
						WatchFields:       nil,
						WatchFieldsAll:    false,
						PassThroughEvents: nil,
					},
					"project/tickets": {},
				},
				Status: common.SubscriptionStatusFailedToRollback,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully rolled back creation of tickets when subscription to contacts failed",
			Input: common.SubscribeParams{
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts":        {Events: eventTypesCUD},
					"project/tickets": {Events: eventTypesCUD},
				},
				Request: &Request{
					WebhookURL: "https://test.com/webhook",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForContacts),
					},
					Then: mockserver.Response(http.StatusBadRequest, errorCreateBadRequest2),
				}, {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks"),
						mockcond.BodyBytes(requestWebhookForTickets),
					},
					Then: mockserver.Response(http.StatusOK, responseWebhookForTickets),
				}, {
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v4_6_release/apis/3.0/system/callbacks/26552"),
					},
					Then: mockserver.Response(http.StatusNoContent),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareResult),
			Expected: &common.SubscriptionResult{
				Result: &Result{},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts":        {}, // Contacts removed successfully.
					"project/tickets": {},
				},
				Status: common.SubscriptionStatusFailed,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionCreator, error) {
				return constructTestStrategy(tt.Server)
			})
		})
	}
}

func compareResult(expectedResult, actualResult *Result) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	if !result.Assert("Result.ObjectWebhooks length",
		len(expectedResult.ObjectWebhooks), len(actualResult.ObjectWebhooks)) {
		return result
	}

	for key, expectedValue := range expectedResult.ObjectWebhooks {
		actualValue, ok := actualResult.ObjectWebhooks[key]
		if !ok {
			actualKeys := make([]string, 0)
			for name := range actualResult.ObjectWebhooks {
				actualKeys = append(actualKeys, name.String())
			}
			result.AddDiff("Result.ObjectWebhooks is missing key [%v], but have (%v)",
				key, strings.Join(actualKeys, ","))

			continue
		}

		result.Assert(fmt.Sprintf("Result.ObjectWebhooks[%v].ID", key), expectedValue.ID, actualValue.ID)
		result.Assert(fmt.Sprintf("Result.ObjectWebhooks[%v].ObjectType", key),
			expectedValue.ObjectType, actualValue.ObjectType)
		result.Assert(fmt.Sprintf("Result.ObjectWebhooks[%v].WebhookURL", key),
			expectedValue.WebhookURL, actualValue.WebhookURL)
	}

	return result
}

func constructTestStrategy(server *httptest.Server) (*Strategy, error) {
	transport, err := components.NewTransport(providers.ConnectWise, common.ConnectorParams{
		AuthenticatedClient: server.Client(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestMockServerBaseURL(server.URL)

	strategy := NewStrategy(transport.JSONHTTPClient(), transport.ProviderInfo(), "test-client-id")

	return strategy, nil
}

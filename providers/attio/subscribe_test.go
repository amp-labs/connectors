package attio

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestSubscribe(t *testing.T) {
	t.Parallel()

	responseObjectsList := testutils.DataFromFile(t, "objects.json")
	responseSubscribeCoreObjects := testutils.DataFromFile(t, "subscribe_core_objects.json")
	responseStandardCustomObjects := testutils.DataFromFile(t, "subscribe_standard_objects.json")

	tests := []testroutines.TestCase[common.SubscribeParams, *common.SubscriptionResult]{
		{
			Name: "Subscribe with missing events",
			Input: common.SubscribeParams{
				Request: &SubscriptionRequest{
					WebhookEndpoint: "https://webbhok.test",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/v2/objects"),
						Then: mockserver.Response(http.StatusOK, responseObjectsList),
					},
				},
			}.Server(),
			ExpectedErrs: []error{errMissingParams},
		},

		{
			Name: "Subscribe with unsupported object",
			Input: common.SubscribeParams{
				Request: &SubscriptionRequest{
					WebhookEndpoint: "https://webbhok.test",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"unsupported_object": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
						},
					},
				},
			},

			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/v2/objects"),
						Then: mockserver.Response(http.StatusOK, responseObjectsList),
					},
				},
			}.Server(),
			ExpectedErrs: []error{fmt.Errorf("unsupported_object: object not found. Ensure it is activated in the workspace settings")},
		},

		{
			Name: "Subscribe only core objects",
			Input: common.SubscribeParams{
				Request: &SubscriptionRequest{
					WebhookEndpoint: "https://webbhok.test",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"lists": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeDelete,
						},
					},

					"tasks": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
						},
					},
				},
			},

			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/v2/objects"),
						Then: mockserver.Response(http.StatusOK, responseObjectsList),
					},
					{
						If:   mockcond.Path("/v2/webhooks"),
						Then: mockserver.Response(http.StatusCreated, responseSubscribeCoreObjects),
					},
				},
			}.Server(),
			ExpectedErrs: nil,
			Comparator: func(_ string, actual, expected *common.SubscriptionResult) bool {
				return actual != nil && actual.Status == common.SubscriptionStatusSuccess
			},
		},

		{
			Name: "Subscribe only standard/custom objects",
			Input: common.SubscribeParams{
				Request: &SubscriptionRequest{
					WebhookEndpoint: "https://webbhok.test",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"people": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeDelete,
						},
					},

					"companies": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
			},

			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.Path("/v2/objects"),
						Then: mockserver.Response(http.StatusOK, responseObjectsList),
					},
					{
						If:   mockcond.Path("/v2/webhooks"),
						Then: mockserver.Response(http.StatusCreated, responseStandardCustomObjects),
					},
				},
			}.Server(),

			Expected: &common.SubscriptionResult{
				Status: common.SubscriptionStatusSuccess,
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"people": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
							common.SubscriptionEventTypeDelete,
						},
					},
					"companies": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
				Result: &subscriptionResult{
					Data: createSubscriptionsResponseData{
						TargetURL: "https://example.com/webhook",
						Status:    "active",
						CreatedAt: "2026-01-30T10:06:11.304000000Z",
						ID: createSubscriptionsResponseID{
							WorkspaceID: "e8d74639-96e5-41be-af46-ced812aef5c5",
							WebhookID:   "c570dd25-5ded-44f6-b94a-84250956455d",
						},
						Secret: "a3ff6435ef497716835413ee10348624a817c87f0f11f7d14e79a68fa5292ebf",
						Subscriptions: []subscription{
							{
								EventType: "record.deleted",
								Filter: map[string]any{
									"$and": []any{
										map[string]any{
											"field":    "id.object_id",
											"operator": "equals",
											"value":    "0e80364d-70b1-44d3-b7ba-0a6a564a7152",
										},
									},
								},
							},
							{
								EventType: "record.updated",
								Filter: map[string]any{
									"$and": []any{
										map[string]any{
											"field":    "id.object_id",
											"operator": "equals",
											"value":    "0e80364d-70b1-44d3-b7ba-0a6a564a7152",
										},
									},
								},
							},
							{
								EventType: "record.created",
								Filter: map[string]any{
									"$and": []any{
										map[string]any{
											"field":    "id.object_id",
											"operator": "equals",
											"value":    "0e80364d-70b1-44d3-b7ba-0a6a564a7152",
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Helper()

			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.Subscribe(t.Context(), tt.Input)

			log.Printf("result: %v:", result)

			tt.Validate(t, err, result)

		})
	}

}

package attio

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

func TestCreateSubscribe(t *testing.T) {
	t.Parallel()

	responseObjectsList := testutils.DataFromFile(t, "objects.json")
	responseSubscribeCoreObjects := testutils.DataFromFile(t, "create_subscribe_core_obj.json")
	responseStandardCustomObjects := testutils.DataFromFile(t, "create_subscribe_standard_obj.json")
	responseCoreStandardObjects := testutils.DataFromFile(t, "create_subscribe_core_standard_obj.json")

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
				Result: &SubscriptionResult{
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

		{
			Name: "Subscribe both core and standard/custom objects",
			Input: common.SubscribeParams{
				Request: &SubscriptionRequest{
					WebhookEndpoint: "https://webbhok.test",
				},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
					"lists": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
						},
					},
					"people": {
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
						Then: mockserver.Response(http.StatusCreated, responseCoreStandardObjects),
					},
				},
			}.Server(),

			Expected: &common.SubscriptionResult{
				Status: common.SubscriptionStatusSuccess,
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"lists": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeCreate,
						},
					},
					"people": {
						Events: []common.SubscriptionEventType{
							common.SubscriptionEventTypeUpdate,
						},
					},
				},
				Result: &SubscriptionResult{
					Data: createSubscriptionsResponseData{
						TargetURL: "https://example.com/webhook",
						ID: createSubscriptionsResponseID{
							WorkspaceID: "e8d74639-96e5-41be-af46-ced812aef5c5",
							WebhookID:   "d1c60c7a-c895-4a4a-ba2f-249aeb359d17",
						},
						Status:    "active",
						CreatedAt: "2026-01-30T13:04:22.051000000Z",
						Secret:    "a7fbac2b0dbdfa5b1e876c22eedcd9c852a24738bd56373ed1d008a49f17bcef",
						Subscriptions: []subscription{
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
								EventType: "list.updated",
								Filter:    nil,
							},
						},
					},
				},
			},
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

			tt.Validate(t, err, result)

		})
	}

}

func TestDeleteSubscribe(t *testing.T) {
	t.Parallel()

	tests := []testroutines.TestCase[common.SubscriptionResult, error]{

		{
			Name:         "Unsubscribe with missing result data",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},

		{
			Name: "Unsubscribe successfully",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Data: createSubscriptionsResponseData{
						TargetURL: "https://example.com/webhook",
						ID: createSubscriptionsResponseID{
							WorkspaceID: "e8d74639-96e5-41be-af46-ced812aef5c5",
							WebhookID:   "d1c60c7a-c895-4a4a-ba2f-249aeb359d17",
						},
						Status:    "active",
						CreatedAt: "2026-01-30T13:04:22.051000000Z",
						Secret:    "a7fbac2b0dbdfa5b1e876c22eedcd9c852a24738bd56373ed1d008a49f17bcef",
						Subscriptions: []subscription{
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
								EventType: "list.updated",
								Filter:    nil,
							},
						},
					},
				},
			},

			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/webhooks/d1c60c7a-c895-4a4a-ba2f-249aeb359d17"),
				Then:  mockserver.Response(http.StatusNoContent, nil),
			}.Server(),

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

			err = conn.DeleteSubscription(t.Context(), tt.Input)

			tt.Validate(t, err, nil)
		})
	}
}

func TestValidationSubscriptionEvents(t *testing.T) {
	t.Parallel()

	standardObjects := map[common.ObjectName]string{
		"people":    "obj_123",
		"companies": "obj_456",
	}

	unsupportedObject := map[common.ObjectName]common.ObjectEvents{
		"unsupported_object": {
			Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
			},
		},
	}

	err := validateSubscriptionEvents(unsupportedObject, standardObjects)
	assert.ErrorContains(t, err, "unsupported_object: object not found. Ensure it is activated in the workspace settings")

	unsupportedStandObjectEvents := map[common.ObjectName]common.ObjectEvents{
		"people": {
			Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeOther,
			},
		},
	}

	err = validateSubscriptionEvents(unsupportedStandObjectEvents, standardObjects)
	assert.ErrorContains(t, err, "unsupported subscription event: 'other' for object 'people")

	unsupportedCoreObjectEvents := map[common.ObjectName]common.ObjectEvents{
		"lists": {
			Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeOther,
			},
		},
	}

	err = validateSubscriptionEvents(unsupportedCoreObjectEvents, standardObjects)
	assert.ErrorContains(t, err, "unsupported subscription event for object 'lists'")

	supportedObjectEvents := map[common.ObjectName]common.ObjectEvents{
		"people": {
			Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeUpdate,
			},
		},
		"lists": {
			Events: []common.SubscriptionEventType{
				common.SubscriptionEventTypeCreate,
				common.SubscriptionEventTypeDelete,
			},
		},
	}

	err = validateSubscriptionEvents(supportedObjectEvents, standardObjects)
	assert.NilError(t, err)

}

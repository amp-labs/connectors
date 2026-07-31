package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"gotest.tools/v3/assert"
)

func TestCreateSubscribe(t *testing.T) {
	t.Parallel()

	responseObjectsList := testutils.DataFromFile(t, "objects.json")
	responseSubscribeCoreObjects := testutils.DataFromFile(t, "create_subscribe_core_obj.json")
	responseStandardCustomObjects := testutils.DataFromFile(t, "create_subscribe_standard_obj.json")
	responseCoreStandardObjects := testutils.DataFromFile(t, "create_subscribe_core_standard_obj.json")

	tests := []testconn.TestCase[common.SubscribeParams, *common.SubscriptionResult]{
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
			ExpectedErrs: []error{errObjectNotFound},
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
			Comparator: func(_ string, actual, expected *common.SubscriptionResult) *testutils.CompareResult {
				result := testutils.NewCompareResult()
				if actual == nil {
					result.AddDiff("actual SubscriptionResult is nil")

					return result
				}

				result.Assert("Status", common.SubscriptionStatusSuccess, actual.Status)

				return result
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
					Data: CreateSubscriptionsResponseData{
						TargetURL: "https://example.com/webhook",
						Status:    "active",
						CreatedAt: "2026-01-30T10:06:11.304000000Z",
						Id: CreateSubscriptionsResponseId{
							WorkspaceId: "e8d74639-96e5-41be-af46-ced812aef5c5",
							WebhookId:   "c570dd25-5ded-44f6-b94a-84250956455d",
						},
						Subscriptions: []Subscription{
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
					Data: CreateSubscriptionsResponseData{
						TargetURL: "https://example.com/webhook",
						Id: CreateSubscriptionsResponseId{
							WorkspaceId: "e8d74639-96e5-41be-af46-ced812aef5c5",
							WebhookId:   "d1c60c7a-c895-4a4a-ba2f-249aeb359d17",
						},
						Status:    "active",
						CreatedAt: "2026-01-30T13:04:22.051000000Z",
						Subscriptions: []Subscription{
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

func TestUpdateSubscribe(t *testing.T) {
	t.Parallel()

	responseObjectsList := testutils.DataFromFile(t, "objects.json")

	// Update response intentionally omits "secret" so we can assert it is preserved from the
	// previous result (the webhook id is unchanged, so the signing secret is unchanged).
	updatedWebhookResponse := []byte(`{
		"data": {
			"target_url": "https://webhook.test",
			"subscriptions": [{"event_type": "list.updated", "filter": null}],
			"id": {"workspace_id": "ws-1", "webhook_id": "wh-1"},
			"status": "active",
			"created_at": "2026-01-01T00:00:00.000000000Z"
		}
	}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If:   mockcond.Path("/v2/objects"),
				Then: mockserver.Response(http.StatusOK, responseObjectsList),
			},
			{
				If:   mockcond.And{mockcond.MethodPATCH(), mockcond.Path("/v2/webhooks/wh-1")},
				Then: mockserver.Response(http.StatusOK, updatedWebhookResponse),
			},
		},
	}.Server()

	t.Cleanup(server.Close)

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	previousResult := &common.SubscriptionResult{
		Result: &SubscriptionResult{
			Data: CreateSubscriptionsResponseData{
				Id:     CreateSubscriptionsResponseId{WebhookId: "wh-1"},
				Secret: "prev-secret",
				Subscriptions: []Subscription{
					{EventType: "list.created"},
				},
			},
		},
	}

	params := common.SubscribeParams{
		Request: &SubscriptionRequest{WebhookEndpoint: "https://webhook.test"},
		SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
			"lists": {
				Events: []common.SubscriptionEventType{common.SubscriptionEventTypeUpdate},
			},
		},
	}

	result, err := conn.UpdateSubscription(t.Context(), params, previousResult)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != common.SubscriptionStatusSuccess {
		t.Fatalf("expected status %q, got %q", common.SubscriptionStatusSuccess, result.Status)
	}

	updated, ok := result.Result.(*SubscriptionResult)
	if !ok {
		t.Fatalf("expected *SubscriptionResult, got %T", result.Result)
	}

	if updated.Data.Secret != "prev-secret" {
		t.Fatalf("expected secret to be preserved as %q, got %q", "prev-secret", updated.Data.Secret)
	}

	// Missing previous result is an error.
	if _, err := conn.UpdateSubscription(t.Context(), params, nil); err == nil {
		t.Fatal("expected error for nil previous result, got nil")
	}

	// Previous result without a webhook id is an error.
	noWebhookID := &common.SubscriptionResult{Result: &SubscriptionResult{}}
	if _, err := conn.UpdateSubscription(t.Context(), params, noWebhookID); err == nil {
		t.Fatal("expected error for missing webhook id, got nil")
	}
}

func TestDeleteSubscribe(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCase[common.SubscriptionResult, error]{
		{
			Name:         "Unsubscribe with missing result data",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},

		{
			Name: "Unsubscribe successfully",
			Input: common.SubscriptionResult{
				Result: &SubscriptionResult{
					Data: CreateSubscriptionsResponseData{
						TargetURL: "https://example.com/webhook",
						Id: CreateSubscriptionsResponseId{
							WorkspaceId: "e8d74639-96e5-41be-af46-ced812aef5c5",
							WebhookId:   "d1c60c7a-c895-4a4a-ba2f-249aeb359d17",
						},
						Status:    "active",
						CreatedAt: "2026-01-30T13:04:22.051000000Z",
						Subscriptions: []Subscription{
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

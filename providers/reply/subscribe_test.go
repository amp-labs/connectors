package reply

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

var (
	_ testconn.TestableSubscriptionCreator = &Connector{}
	_ testconn.TestableSubscriptionUpdater = &Connector{}
	_ testconn.TestableSubscriptionRemover = &Connector{}
)

const (
	hookURL      = "https://hooks.example.com/reply/installation-1"
	otherHookURL = "https://hooks.example.com/reply/installation-2"

	limitReached = `{"code":"webHook.limitReached","title":"Conflict","status":409,` +
		`"detail":"Limit of active webhook subscriptions reached (50)."}`
)

func webhookBody(id int, event, url string) []byte {
	return fmt.Appendf(nil, `{"id":%d,"eventType":%q,"url":%q,"scope":"personal","enabled":true,`+
		`"payloadConfig":{"includeEmailUrl":false,"includeEmailText":false,`+
		`"includeLinkedInMessageText":false,"includeProspectCustomFields":false}}`, id, event, url)
}

// createsWebhook answers the POST that subscribes to the event.
func createsWebhook(event, url string, id int) mockserver.Case {
	return mockserver.Case{
		If: mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/v3/webhooks"),
			mockcond.Body(fmt.Sprintf(`{"eventType":%q,"url":%q,"scope":"personal"}`, event, url)),
		},
		Then: mockserver.Response(http.StatusCreated, webhookBody(id, event, url)),
	}
}

func rejectsWebhook(event, url string) mockserver.Case {
	return mockserver.Case{
		If: mockcond.And{
			mockcond.MethodPOST(),
			mockcond.Path("/v3/webhooks"),
			mockcond.Body(fmt.Sprintf(`{"eventType":%q,"url":%q,"scope":"personal"}`, event, url)),
		},
		Then: mockserver.Response(http.StatusConflict, []byte(limitReached)),
	}
}

// deletesWebhook answers the DELETE of the subscription with the given status.
func deletesWebhook(id, status int) mockserver.Case {
	return mockserver.Case{
		If: mockcond.And{
			mockcond.MethodDELETE(),
			mockcond.Path(fmt.Sprintf("/v3/webhooks/%d", id)),
		},
		Then: mockserver.Response(status),
	}
}

func switchOf(cases ...mockserver.Case) *mockserver.Switch {
	return &mockserver.Switch{Setup: mockserver.ContentJSON(), Cases: cases}
}

func subscribeParams(events map[common.ObjectName]common.ObjectEvents) common.SubscribeParams {
	return common.SubscribeParams{
		Request:            &SubscriptionRequest{WebhookURL: hookURL},
		SubscriptionEvents: events,
	}
}

func passThrough(events ...string) common.ObjectEvents {
	return common.ObjectEvents{PassThroughEvents: events}
}

func hook(id int, event, url string) WebhookSubscription {
	return WebhookSubscription{ID: id, EventType: event, URL: url, Scope: "personal"}
}

func previousResult(subscriptions ...WebhookSubscription) *common.SubscriptionResult {
	return subscriptionResult(subscriptions, common.SubscriptionStatusSuccess)
}

func compareSubscriptions(expected, actual *SubscriptionResult) *testutils.CompareResult {
	result := testutils.NewCompareResult()
	result.Assert("Subscriptions", expected.Subscriptions, actual.Subscriptions)

	return result
}

func TestSubscribe(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseSubscribe{
		{
			Name:         "Request must be provided",
			Input:        common.SubscribeParams{SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{"contacts": passThrough("contact_replied")}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingSubscribeParams},
		},
		{
			Name: "Webhook URL must be a URL",
			Input: common.SubscribeParams{
				Request:            &SubscriptionRequest{WebhookURL: "not a url"},
				SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{"contacts": passThrough("contact_replied")},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errInvalidSubscribeRequest},
		},
		{
			Name:         "At least one event must be requested",
			Input:        subscribeParams(map[common.ObjectName]common.ObjectEvents{"contacts": {}}),
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errNoEventsRequested},
		},
		{
			Name: "Create, update and delete events are not offered by Reply",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts": {Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate}},
			}),
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errCRUDEventsNotSupported},
		},
		{
			Name: "An event Reply does not offer is rejected",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts": passThrough("email_clicked"),
			}),
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errUnknownPassThroughEvent},
		},
		{
			Name: "An event requested under another object is rejected",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts": passThrough("autopilot_stopped"),
			}),
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errEventObjectMismatch},
		},
		{
			Name: "One subscription is created per event, in a stable order",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts":  passThrough("email_replied", "contact_opted_out", "email_replied"),
				"sequences": passThrough("autopilot_stopped"),
			}),
			Server: switchOf(
				createsWebhook("autopilot_stopped", hookURL, 103),
				createsWebhook("contact_opted_out", hookURL, 101),
				createsWebhook("email_replied", hookURL, 102),
			).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected: &common.SubscriptionResult{
				Status: common.SubscriptionStatusSuccess,
				Result: &SubscriptionResult{Subscriptions: []WebhookSubscription{
					hook(103, "autopilot_stopped", hookURL),
					hook(101, "contact_opted_out", hookURL),
					hook(102, "email_replied", hookURL),
				}},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"contacts":  passThrough("contact_opted_out", "email_replied"),
					"sequences": passThrough("autopilot_stopped"),
				},
			},
		},
		{
			Name: "A failed creation deletes the subscriptions already created",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts":  passThrough("contact_opted_out"),
				"sequences": passThrough("autopilot_stopped"),
			}),
			Server: switchOf(
				createsWebhook("autopilot_stopped", hookURL, 201),
				rejectsWebhook("contact_opted_out", hookURL),
				deletesWebhook(201, http.StatusNoContent),
			).Server(),
			Comparator:   testconn.ComparatorSubscriptionWithoutResult,
			Expected:     &common.SubscriptionResult{Status: common.SubscriptionStatusFailed},
			ExpectedErrs: []error{testutils.StringError("webHook.limitReached")},
		},
		{
			Name: "A creation that fails first leaves nothing to roll back",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts": passThrough("contact_opted_out"),
			}),
			Server:       switchOf(rejectsWebhook("contact_opted_out", hookURL)).Server(),
			Comparator:   testconn.ComparatorSubscriptionWithoutResult,
			Expected:     &common.SubscriptionResult{Status: common.SubscriptionStatusFailed},
			ExpectedErrs: []error{testutils.StringError("webHook.limitReached")},
		},
		{
			Name: "A subscription that cannot be rolled back stays in the result",
			Input: subscribeParams(map[common.ObjectName]common.ObjectEvents{
				"contacts":  passThrough("contact_opted_out"),
				"sequences": passThrough("autopilot_stopped"),
			}),
			Server: switchOf(
				createsWebhook("autopilot_stopped", hookURL, 201),
				rejectsWebhook("contact_opted_out", hookURL),
				deletesWebhook(201, http.StatusInternalServerError),
			).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected: &common.SubscriptionResult{
				Status: common.SubscriptionStatusFailedToRollback,
				Result: &SubscriptionResult{Subscriptions: []WebhookSubscription{
					hook(201, "autopilot_stopped", hookURL),
				}},
				ObjectEvents: map[common.ObjectName]common.ObjectEvents{
					"sequences": passThrough("autopilot_stopped"),
				},
			},
			ExpectedErrs: []error{testutils.StringError("webHook.limitReached"), common.ErrServer},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionCreator, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestUpdateSubscription(t *testing.T) { //nolint:funlen
	t.Parallel()

	contactsAndReplies := previousResult(
		hook(11, "contact_opted_out", hookURL),
		hook(12, "email_replied", hookURL),
	)

	tests := []testconn.TestCaseUpdateSubscription{
		{
			Name: "The previous result must be provided",
			Input: testconn.UpdateSubscriptionParams{
				Params: subscribeParams(map[common.ObjectName]common.ObjectEvents{
					"contacts": passThrough("contact_replied"),
				}),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errInvalidSubscriptionResult},
		},
		{
			Name: "Unchanged events call Reply for nothing",
			Input: testconn.UpdateSubscriptionParams{
				Params: subscribeParams(map[common.ObjectName]common.ObjectEvents{
					"contacts": passThrough("contact_opted_out", "email_replied"),
				}),
				PreviousResult: contactsAndReplies,
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected:   contactsAndReplies,
		},
		{
			Name: "New events are created and events no longer wanted are deleted",
			Input: testconn.UpdateSubscriptionParams{
				Params: subscribeParams(map[common.ObjectName]common.ObjectEvents{
					"contacts": passThrough("contact_opted_out", "contact_finished"),
				}),
				PreviousResult: contactsAndReplies,
			},
			Server: switchOf(
				createsWebhook("contact_finished", hookURL, 13),
				deletesWebhook(12, http.StatusNoContent),
			).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected: previousResult(
				hook(13, "contact_finished", hookURL),
				hook(11, "contact_opted_out", hookURL),
			),
		},
		{
			Name: "An event whose URL changed is subscribed again at the new URL",
			Input: testconn.UpdateSubscriptionParams{
				Params: common.SubscribeParams{
					Request: &SubscriptionRequest{WebhookURL: otherHookURL},
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"contacts": passThrough("contact_opted_out"),
					},
				},
				PreviousResult: previousResult(hook(11, "contact_opted_out", hookURL)),
			},
			Server: switchOf(
				createsWebhook("contact_opted_out", otherHookURL, 14),
				deletesWebhook(11, http.StatusNoContent),
			).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected:   previousResult(hook(14, "contact_opted_out", otherHookURL)),
		},
		{
			Name: "A subscription that is already gone counts as deleted",
			Input: testconn.UpdateSubscriptionParams{
				Params: subscribeParams(map[common.ObjectName]common.ObjectEvents{
					"contacts": passThrough("contact_opted_out"),
				}),
				PreviousResult: contactsAndReplies,
			},
			Server:     switchOf(deletesWebhook(12, http.StatusNotFound)).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected:   previousResult(hook(11, "contact_opted_out", hookURL)),
		},
		{
			Name: "A failed creation keeps the previous state and creates nothing else",
			Input: testconn.UpdateSubscriptionParams{
				Params: subscribeParams(map[common.ObjectName]common.ObjectEvents{
					"contacts": passThrough("contact_opted_out", "email_replied", "contact_finished"),
				}),
				PreviousResult: contactsAndReplies,
			},
			Server:     switchOf(rejectsWebhook("contact_finished", hookURL)).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected: subscriptionResult([]WebhookSubscription{
				hook(11, "contact_opted_out", hookURL),
				hook(12, "email_replied", hookURL),
			}, common.SubscriptionStatusFailed),
			ExpectedErrs: []error{testutils.StringError("webHook.limitReached")},
		},
		{
			Name: "A subscription that cannot be deleted stays in the result",
			Input: testconn.UpdateSubscriptionParams{
				Params: subscribeParams(map[common.ObjectName]common.ObjectEvents{
					"contacts": passThrough("contact_opted_out", "contact_finished"),
				}),
				PreviousResult: contactsAndReplies,
			},
			Server: switchOf(
				createsWebhook("contact_finished", hookURL, 13),
				deletesWebhook(12, http.StatusInternalServerError),
			).Server(),
			Comparator: testconn.ComparatorSubscriptionWithResult(compareSubscriptions),
			Expected: subscriptionResult([]WebhookSubscription{
				hook(13, "contact_finished", hookURL),
				hook(11, "contact_opted_out", hookURL),
				hook(12, "email_replied", hookURL),
			}, common.SubscriptionStatusFailedToRollback),
			ExpectedErrs: []error{common.ErrServer},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionUpdater, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	t.Parallel()

	two := previousResult(hook(11, "contact_opted_out", hookURL), hook(12, "email_replied", hookURL))

	tests := []testconn.TestCaseDeleteSubscription{
		{
			Name:         "The result must be provided",
			Input:        common.SubscriptionResult{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errInvalidSubscriptionResult},
		},
		{
			Name:   "Nothing to delete",
			Input:  common.SubscriptionResult{Result: &SubscriptionResult{}},
			Server: mockserver.Dummy(),
		},
		{
			Name:  "Every subscription is deleted",
			Input: *two,
			Server: switchOf(
				deletesWebhook(11, http.StatusNoContent),
				deletesWebhook(12, http.StatusNoContent),
			).Server(),
		},
		{
			Name:  "A subscription that is already gone counts as deleted",
			Input: *two,
			Server: switchOf(
				deletesWebhook(11, http.StatusNotFound),
				deletesWebhook(12, http.StatusNoContent),
			).Server(),
		},
		{
			Name:  "A failure does not stop the others from being deleted",
			Input: *two,
			Server: switchOf(
				deletesWebhook(11, http.StatusInternalServerError),
				deletesWebhook(12, http.StatusNoContent),
			).Server(),
			ExpectedErrs: []error{common.ErrServer},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSubscriptionRemover, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

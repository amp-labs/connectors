package google

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
)

// TestMailSubscribeWatchRequest checks the users.watch body built for each set of
// subscribed objects. "messages" must not be filtered by label: sent mail and received
// mail that skips the inbox (archived by a filter, spam) never carry INBOX, so an
// INBOX filter would never push for them.
func TestMailSubscribeWatchRequest(t *testing.T) { //nolint:funlen
	t.Parallel()

	const (
		topic     = "projects/test-project/topics/gmail-subscribe"
		watchPath = "/gmail/v1/users/me/watch"
		response  = `{"historyId":"12345","expiration":"1792171812000"}`
	)

	createOnly := common.ObjectEvents{
		Events: []common.SubscriptionEventType{common.SubscriptionEventTypeCreate},
	}

	tests := []struct {
		name         string
		objects      map[common.ObjectName]common.ObjectEvents
		request      *GmailSubscribeRequest
		expectedBody string
		expectErr    bool
	}{
		{
			name:         "messages watches the whole mailbox",
			objects:      map[common.ObjectName]common.ObjectEvents{"messages": createOnly},
			request:      &GmailSubscribeRequest{TopicName: topic},
			expectedBody: `{"topicName":"` + topic + `"}`,
		},
		{
			name: "messages with drafts still watches the whole mailbox",
			objects: map[common.ObjectName]common.ObjectEvents{
				"messages": createOnly,
				"drafts":   createOnly,
			},
			request:      &GmailSubscribeRequest{TopicName: topic},
			expectedBody: `{"topicName":"` + topic + `"}`,
		},
		{
			name:         "drafts alone filters on the DRAFT label",
			objects:      map[common.ObjectName]common.ObjectEvents{"drafts": createOnly},
			request:      &GmailSubscribeRequest{TopicName: topic},
			expectedBody: `{"topicName":"` + topic + `","labelIds":["DRAFT"],"labelFilterBehavior":"include"}`,
		},
		{
			name:    "explicit label override is kept",
			objects: map[common.ObjectName]common.ObjectEvents{"messages": createOnly},
			request: &GmailSubscribeRequest{
				TopicName:           topic,
				LabelIDs:            []string{"INBOX"},
				LabelFilterBehavior: "include",
			},
			expectedBody: `{"topicName":"` + topic + `","labelIds":["INBOX"],"labelFilterBehavior":"include"}`,
		},
		{
			name:      "unsupported object is rejected",
			objects:   map[common.ObjectName]common.ObjectEvents{"threads": createOnly},
			request:   &GmailSubscribeRequest{TopicName: topic},
			expectErr: true,
		},
		{
			name: "unsupported object is rejected alongside messages",
			objects: map[common.ObjectName]common.ObjectEvents{
				"messages": createOnly,
				"threads":  createOnly,
			},
			request:   &GmailSubscribeRequest{TopicName: topic},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path(watchPath),
					mockcond.Body(tt.expectedBody),
				},
				Then: mockserver.Response(http.StatusOK, []byte(response)),
			}.Server()
			t.Cleanup(srv.Close)

			conn, err := constructTestMailConnector(srv.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.Subscribe(t.Context(), common.SubscribeParams{
				Request:            tt.request,
				SubscriptionEvents: tt.objects,
			})

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error (watch body did not match %s?): %v", tt.expectedBody, err)
			}

			if result.Status != common.SubscriptionStatusSuccess {
				t.Fatalf("expected status %q, got %q", common.SubscriptionStatusSuccess, result.Status)
			}
		})
	}
}

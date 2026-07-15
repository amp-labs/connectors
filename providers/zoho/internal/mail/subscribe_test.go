package mail

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"gotest.tools/v3/assert"
)

// testHookSecret is an arbitrary HMAC key; the tests sign their own bodies with
// it, so the value is not sensitive — kept low-entropy to avoid secret scanners.
const testHookSecret = "unit-test-hook-secret"

// mailWebhookBody is a Zoho Mail outgoing-webhook payload for a new email, from
// the WEBHOOK RESPONSE SAMPLE in the docs. messageId and folderId are 64-bit
// integers larger than 2^53.
const mailWebhookBody = `{
	"summary": "Hi Rebecca, please take a look.",
	"sentDateInGMT": 1560866021000,
	"subject": "Marketing - Product pitch",
	"messageId": 1560840837125110000,
	"toAddress": "\"Rebecca A\"<rebecca@zylker.com>",
	"folderId": 3881227000000013000,
	"zuid": 647772765,
	"ccAddress": "",
	"size": 55503,
	"sender": "Paula",
	"receivedTime": 1560840837126,
	"fromAddress": "paula@zylker.com",
	"html": "<div>Hi Rebecca,</div>",
	"IntegIdList": "34000000580271,"
}`

func signBody(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func newTestAdapter(t *testing.T, secret string) *Adapter {
	t.Helper()

	adapter, err := NewAdapter(&common.JSONHTTPClient{}, &providers.ModuleInfo{BaseURL: "https://mail.zoho.com"}, "", secret)
	if err != nil {
		t.Fatalf("failed to construct adapter: %v", err)
	}

	return adapter
}

func TestVerifyWebhookMessage(t *testing.T) { //nolint:funlen
	t.Parallel()

	body := []byte(mailWebhookBody)
	validSig := signBody(testHookSecret, mailWebhookBody)

	tests := []struct {
		name         string
		secret       string
		headers      http.Header
		expected     bool
		expectedErrs []error
	}{
		{
			name:         "Missing secret rejects every message",
			secret:       "",
			headers:      http.Header{mailHookSignatureHeader: []string{validSig}},
			expected:     false,
			expectedErrs: []error{ErrMissingWebhookSecret},
		},
		{
			name:         "Missing signature header",
			secret:       testHookSecret,
			headers:      http.Header{},
			expected:     false,
			expectedErrs: []error{common.ErrMissingHeader},
		},
		{
			name:     "Invalid signature",
			secret:   testHookSecret,
			headers:  http.Header{mailHookSignatureHeader: []string{"not-the-right-signature"}},
			expected: false,
		},
		{
			name:     "Valid signature",
			secret:   testHookSecret,
			headers:  http.Header{mailHookSignatureHeader: []string{validSig}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := newTestAdapter(t, tt.secret)

			ok, err := adapter.VerifyWebhookMessage(context.Background(),
				&common.WebhookRequest{Headers: tt.headers, Body: body},
				&common.VerificationParams{},
			)

			assert.Equal(t, ok, tt.expected)

			for _, wantErr := range tt.expectedErrs {
				assert.ErrorIs(t, err, wantErr)
			}

			if len(tt.expectedErrs) == 0 && !tt.expected {
				assert.NilError(t, err) // invalid signature: mismatch, but no error
			}
		})
	}
}

// TestMailSubscriptionEvent verifies the payload parses correctly, including
// preserving the large messageId (decoded via json.Number, no precision loss).
func TestMailSubscriptionEvent(t *testing.T) { //nolint:funlen
	t.Parallel()

	var evt *CollapsedSubscriptionEvent

	dec := json.NewDecoder(bytes.NewReader([]byte(mailWebhookBody)))
	dec.UseNumber() // keep 64-bit ids exact; do not let them become float64

	if err := dec.Decode(&evt); err != nil {
		t.Fatalf("failed to unmarshal mail event: %v", err)
	}

	assert.Equal(t, IsWebhookPayload(*evt), true)

	subevts, err := evt.SubscriptionEventList()
	assert.NilError(t, err)
	assert.Equal(t, len(subevts), 1) // Zoho Mail sends one email per webhook

	subevt := subevts[0]

	eventType, err := subevt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, eventType, common.SubscriptionEventTypeCreate)

	rawEventName, err := subevt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, rawEventName, rawNameNewMail)

	objectName, err := subevt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, objectName, objectNameMessages)

	recordID, err := subevt.RecordId()
	assert.NilError(t, err)
	// composite "<folderId>/<messageId>", both exact (> 2^53)
	assert.Equal(t, recordID, "3881227000000013000/1560840837125110000")

	tsNano, err := subevt.EventTimeStampNano()
	assert.NilError(t, err)
	assert.Equal(t, tsNano, int64(1560840837126)*int64(1_000_000))

	workspace, err := subevt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, workspace, "647772765")

	updateEvt, ok := subevt.(common.SubscriptionUpdateEvent)
	assert.Equal(t, ok, true)

	updatedFields, err := updateEvt.UpdatedFields()
	assert.NilError(t, err)
	assert.Equal(t, len(updatedFields), 0)
}

// TestGetRecordsByIds verifies the composite record id is split back into
// folderId + messageId to fetch the message details, and that the unsupported
// object case is rejected.
func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	const (
		folderID  = "3881227000000013000"
		messageID = "1560840837125110000"
	)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If: mockcond.Path(
				"/api/accounts/" + testAccountID + "/folders/" + folderID + "/messages/" + messageID + "/details",
			),
			Then: mockserver.ResponseString(http.StatusOK, `{
				"status": {"code": 200, "description": "success"},
				"data": {"subject": "Marketing - Product pitch", "fromAddress": "paula@zylker.com"}
			}`),
		}},
	}.Server()
	t.Cleanup(server.Close)

	adapter := constructTestAdapter(t, server.URL, testAccountID)

	rows, err := adapter.GetRecordsByIds(context.Background(), objectNameMessages,
		[]string{folderID + "/" + messageID}, connectors.Fields("subject").List(), nil)
	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, rows[0].Id, folderID+"/"+messageID)
	assert.Equal(t, rows[0].Fields["subject"], "Marketing - Product pitch")
	assert.Equal(t, rows[0].Raw["fromAddress"], "paula@zylker.com")

	// Unsupported object is rejected.
	_, err = adapter.GetRecordsByIds(context.Background(), "notes", []string{"1/2"}, nil, nil)
	assert.ErrorIs(t, err, common.ErrGetRecordNotSupportedForObject)

	// Empty recordIds is rejected.
	_, err = adapter.GetRecordsByIds(context.Background(), objectNameMessages, nil, nil, nil)
	assert.ErrorIs(t, err, errNoRecordIDs)

	// Malformed composite id (no folderId) is rejected for messages.
	_, err = adapter.GetRecordsByIds(context.Background(), objectNameMessages, []string{"no-separator"}, nil, nil)
	assert.ErrorIs(t, err, errInvalidRecordID)
}

// taskWebhookBody is a Zoho Mail Task outgoing-webhook payload. entityId and
// nameSpaceId are the task id and group id.
const taskWebhookBody = `{
	"entityId": 4000000012345,
	"entityType": 3,
	"action": "taskUpdated",
	"title": "Prepare pitch deck",
	"summary": "Draft the Q3 pitch",
	"assignee": 647772765,
	"assigneeName": "Rebecca",
	"dueDate": 1560866021000,
	"nameSpaceId": 9000000002014,
	"groupName": "Marketing",
	"status": 2,
	"statusName": "In Progress",
	"triggerZuid": 647772765
}`

// TestTaskSubscriptionEvent verifies a Task webhook payload parses correctly and
// routes to the tasks object with a group-scoped composite record id.
func TestTaskSubscriptionEvent(t *testing.T) { //nolint:funlen
	t.Parallel()

	var evt *CollapsedSubscriptionEvent

	dec := json.NewDecoder(bytes.NewReader([]byte(taskWebhookBody)))
	dec.UseNumber()

	if err := dec.Decode(&evt); err != nil {
		t.Fatalf("failed to unmarshal task event: %v", err)
	}

	assert.Equal(t, IsWebhookPayload(*evt), true)

	subevts, err := evt.SubscriptionEventList()
	assert.NilError(t, err)
	assert.Equal(t, len(subevts), 1)

	subevt := subevts[0]

	objectName, err := subevt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, objectName, objectNameTasks)

	eventType, err := subevt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, eventType, common.SubscriptionEventTypeUpdate) // "taskUpdated"

	rawEventName, err := subevt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, rawEventName, "taskUpdated")

	recordID, err := subevt.RecordId()
	assert.NilError(t, err)
	// composite "<groupId>/<taskId>"
	assert.Equal(t, recordID, "9000000002014/4000000012345")

	workspace, err := subevt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, workspace, "9000000002014") // group id
}

// TestMailRecordIDWithoutUseNumber ensures the lossless messageIdString twin is
// preferred, so the record id survives a plain json.Unmarshal decode (which
// turns the numeric messageId into a float64 and rounds it past 2^53).
func TestMailRecordIDWithoutUseNumber(t *testing.T) {
	t.Parallel()

	var evt CollapsedSubscriptionEvent

	body := `{"messageId":1784105487965154200,"messageIdString":"1784105487965154200","folderId":123}`
	if err := json.Unmarshal([]byte(body), &evt); err != nil { // no UseNumber on purpose
		t.Fatalf("failed to unmarshal mail event: %v", err)
	}

	subevts, err := evt.SubscriptionEventList()
	assert.NilError(t, err)

	recordID, err := subevts[0].RecordId()
	assert.NilError(t, err)
	assert.Equal(t, recordID, "123/1784105487965154200") // exact despite float64 decode
}

// TestIsWebhookPayloadRejectsCRMShape ensures a payload carrying CRM
// discriminator keys is never claimed by the Mail parser, even if it also has
// a messageId/entityId field.
func TestIsWebhookPayloadRejectsCRMShape(t *testing.T) {
	t.Parallel()

	crmWithEntityID := map[string]any{
		crmKeyModule:         "Leads",
		crmKeyOperation:      "update",
		crmKeyAffectedValues: []any{},
		keyEntityID:          "123",
	}
	assert.Equal(t, IsWebhookPayload(crmWithEntityID), false)

	pureCRM := map[string]any{crmKeyModule: "Leads", crmKeyOperation: "update"}
	assert.Equal(t, IsWebhookPayload(pureCRM), false)
}

// TestTaskEventTypeMapping spot-checks the best-effort action -> type mapping.
func TestTaskEventTypeMapping(t *testing.T) {
	t.Parallel()

	cases := map[string]common.SubscriptionEventType{
		"taskAdded":     common.SubscriptionEventTypeCreate,
		"taskCreated":   common.SubscriptionEventTypeCreate,
		"taskUpdated":   common.SubscriptionEventTypeUpdate,
		"taskCompleted": common.SubscriptionEventTypeUpdate,
		"statusChanged": common.SubscriptionEventTypeUpdate,
		"taskDeleted":   common.SubscriptionEventTypeDelete,
		"somethingElse": common.SubscriptionEventTypeOther,
	}

	for action, want := range cases {
		assert.Equal(t, taskEventType(action), want)
	}
}

// TestGetTaskRecordsByIds verifies a group task is fetched from the group
// endpoint and a personal task (bare id) from the personal endpoint.
func TestGetTaskRecordsByIds(t *testing.T) {
	t.Parallel()

	const (
		groupID = "9000000002014"
		taskID  = "4000000012345"
	)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.Path("/api/tasks/groups/" + groupID + "/" + taskID),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"status": {"code": 200, "description": "success"},
					"data": {"tasks": [{"id": "`+taskID+`", "title": "Prepare pitch deck", "status": "2"}]}
				}`),
			},
			{
				If: mockcond.Path("/api/tasks/me/" + taskID),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"status": {"code": 200, "description": "success"},
					"data": {"tasks": [{"id": "`+taskID+`", "title": "Personal task"}]}
				}`),
			},
		},
	}.Server()
	t.Cleanup(server.Close)

	adapter := constructTestAdapter(t, server.URL, testAccountID)

	// Group task: composite "<groupId>/<taskId>".
	rows, err := adapter.GetRecordsByIds(context.Background(), objectNameTasks,
		[]string{groupID + "/" + taskID}, connectors.Fields("title").List(), nil)
	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, rows[0].Id, groupID+"/"+taskID)
	assert.Equal(t, rows[0].Fields["title"], "Prepare pitch deck")

	// Personal task: bare "<taskId>".
	rows, err = adapter.GetRecordsByIds(context.Background(), objectNameTasks,
		[]string{taskID}, connectors.Fields("title").List(), nil)
	assert.NilError(t, err)
	assert.Equal(t, len(rows), 1)
	assert.Equal(t, rows[0].Id, taskID)
	assert.Equal(t, rows[0].Fields["title"], "Personal task")
}

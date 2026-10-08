// Command webhook exercises the Zoho Mail webhook-verification, event-parsing,
// and (optionally) record-enrichment paths end to end, for both the Mail and
// Task webhook entities.
//
// Zoho Mail has NO API to create/update/delete webhook subscriptions — the
// outgoing webhook is configured by hand in the Zoho Mail console
// (Settings > Integrations > Developer Space > Outgoing Webhooks). The
// connector's webhook responsibilities are: verify the HMAC-SHA256 signature
// Zoho attaches to each delivery, parse the payload into events, and (for
// enrichment) fetch full records by id. This harness drives all three.
//
// Usage:
//
//	go run ./test/zoho/mail/webhook
//
// When the webhook* vars below are left empty it runs a self-signed smoke test
// over the built-in Mail and Task sample bodies (verify + parse; enrichment
// skipped, since the sample ids are not real). To exercise a REAL delivery
// captured from Zoho:
//  1. Configure an outgoing Mail or Task webhook in the Zoho Mail console.
//  2. Capture the x-hook-secret from the header of the FIRST request (Zoho
//     sends it only once) plus a later request's raw body + x-hook-signature.
//  3. Fill in the webhookSecret / webhookBody / webhookSignature vars below
//     (and set enrich = true to also fetch full records), then re-run.
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoho"
	"github.com/amp-labs/connectors/test/utils"
	connTest "github.com/amp-labs/connectors/test/zoho"
)

// mailSampleBody is a Zoho Mail "new email" webhook payload (WEBHOOK RESPONSE
// SAMPLE from the docs). taskSampleBody is a Task webhook payload. Both are used
// in self-signed mode when no real captured body is supplied.
const (
	mailSampleBody = `{
		"summary": "Hi Rebecca, please take a look.",
		"sentDateInGMT": 1560866021000,
		"subject": "Marketing - Product pitch",
		"messageId": 1560840837125110000,
		"toAddress": "\"Rebecca A\"<rebecca@zylker.com>",
		"folderId": 3881227000000013000,
		"zuid": 647772765,
		"sender": "Paula",
		"receivedTime": 1560840837126,
		"fromAddress": "paula@zylker.com",
		"html": "<div>Hi Rebecca,</div>"
	}`

	taskSampleBody = `{
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
)

const signatureHeader = "X-Hook-Signature"

// capturedSampleName marks the sample built from webhookBody; the captured
// webhookSignature applies only to that sample, never to the built-in ones.
const capturedSampleName = "captured"

// Fill these in with values captured from a real Zoho Mail webhook delivery.
// Leave them at their defaults to run the built-in self-signed smoke test.
var (
	// webhookSecret is the x-hook-secret Zoho sends on the first webhook request.
	// Empty => a demo secret is used and each sample body is self-signed.
	webhookSecret = "f6f575ca-ff12-43dc-b64c-51f7eda82616"

	// webhookSignature is a captured x-hook-signature for webhookBody. Empty =>
	// the body is self-signed with webhookSecret so the "valid" check passes.
	webhookSignature = "Od/WUg1rQ7GoyOuPFTPRRSShpIV7wI15pudCWo0zJy0="

	// webhookBody is the EXACT raw request body captured from a real Zoho Mail
	// delivery (not a reconstruction) — its bytes reproduce webhookSignature
	// under webhookSecret. Note the wire format that a parsed view hides:
	// forward slashes are escaped (<\/div>), the newline is a literal \r\n, and
	// folderId ends in ...008 (the viewer rounds it to ...000).
	webhookBody = `{"summary":"How are you?","sentDateInGMT":1784130673000,"subject":"Test 3","Mode":0,"messageId":1784105487965154200,"toAddress":"<untergration.user@zohomail.com>","folderId":7214453000000008008,"zuid":848632206,"threadId":0,"hasAttachment":"No","size":255,"sender":"Karage pep","receivedTime":1784105487962,"fromAddress":"pepkarage@gmail.com","html":"<div><div dir=\"ltr\">How are you?<div><br><\/div><\/div>\r\n<\/div>","messageIdString":"1784105487965154200","IntegIdList":"1784105384068156900,1784026830688156900,1784026808883156900,"}`

	// enrich, when true, also fetches each event's full record via
	// GetRecordsByIds. Only meaningful for a real body whose ids exist.
	enrich = true

	// webhookAccountID, if set, is used as the Zoho Mail account id for
	// enrichment, skipping the /api/accounts lookup (which needs the
	// ZohoMail.accounts scope). Leave empty to resolve it via that lookup.
	webhookAccountID = ""
)

type sample struct {
	name string
	body []byte
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	// The webhook secret is the x-hook-secret Zoho delivers on the first
	// request. Fall back to a demo secret so the harness is runnable without a
	// real Zoho setup (self-signed mode).
	secret := webhookSecret
	if secret == "" {
		secret = "demo-secret-only-for-local-self-signing"
		slog.Warn("webhookSecret not set; running in self-signed mode with a demo secret")
	}

	// The connector reads the secret (and optional account id) from connection
	// metadata. Use the no-refresh variant so the creds access token (minted
	// with Zoho Mail scopes) is sent verbatim rather than refreshed into a
	// differently-scoped token.
	metadata := map[string]string{"zohoMailWebhookSecret": secret}
	if webhookAccountID != "" {
		metadata["zohoMailAccountId"] = webhookAccountID
	}

	conn := connTest.GetZohoConnectorNoRefresh(ctx, providers.ModuleZohoMail, metadata)

	// Enrichment hits account-scoped endpoints (api/accounts/{accountId}/...).
	// Resolve the account id via the accounts API only when it wasn't supplied
	// directly — that lookup needs the ZohoMail.accounts scope, which supplying
	// webhookAccountID avoids. Verify + parse never need it.
	if enrich && webhookAccountID == "" {
		if _, err := conn.GetPostAuthInfo(ctx); err != nil {
			utils.Fail("post-authentication (account id resolution) failed", "error", err)
		}
	}

	for _, s := range loadSamples() {
		slog.Info("=== sample ===", "name", s.name)

		// Built-in samples are always self-signed. The captured signature is
		// applied only to the captured body it belongs to (which is likewise
		// self-signed when no signature was captured).
		signature := sign(secret, s.body)
		if s.name == capturedSampleName && webhookSignature != "" {
			signature = webhookSignature
		}

		// 1) A correctly-signed request must verify.
		verify(ctx, conn, s.body, signature, true)

		// 2) A tampered body must fail verification.
		verify(ctx, conn, append(bytes.Clone(s.body), ' '), signature, false)

		// 3) Parse the payload into subscription events, and optionally enrich.
		events := parseEvents(s.body)
		if enrich {
			enrichEvents(ctx, conn, events)
		}
	}

	slog.Info("Zoho Mail webhook harness completed successfully.")
}

// loadSamples returns the captured real body when webhookBody is set, otherwise
// the built-in Mail and Task samples.
func loadSamples() []sample {
	if webhookBody != "" {
		return []sample{{name: capturedSampleName, body: []byte(webhookBody)}}
	}

	return []sample{
		{name: "mail", body: []byte(mailSampleBody)},
		{name: "task", body: []byte(taskSampleBody)},
	}
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func verify(ctx context.Context, conn *zoho.Connector, body []byte, signature string, want bool) {
	ok, err := conn.VerifyWebhookMessage(ctx,
		&common.WebhookRequest{
			Headers: http.Header{signatureHeader: []string{signature}},
			Body:    body,
		},
		&common.VerificationParams{},
	)
	if err != nil {
		utils.Fail("verification returned an error", "error", err)
	}

	if ok != want {
		utils.Fail("unexpected verification result", "got", ok, "want", want)
	}

	slog.Info("verification result as expected", "valid", ok)
}

func parseEvents(body []byte) []common.SubscriptionEvent {
	var collapsed *zoho.CollapsedSubscriptionEvent

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // keep the 64-bit ids exact

	if err := dec.Decode(&collapsed); err != nil {
		utils.Fail("error decoding webhook body", "error", err)
	}

	events, err := collapsed.SubscriptionEventList()
	if err != nil {
		utils.Fail("error building subscription events", "error", err)
	}

	for _, evt := range events {
		objectName, _ := evt.ObjectName()
		eventType, _ := evt.EventType()
		rawName, _ := evt.RawEventName()
		recordID, _ := evt.RecordId()
		tsNano, _ := evt.EventTimeStampNano()

		slog.Info("event",
			"object", objectName,
			"type", eventType,
			"rawName", rawName,
			"recordId", recordID,
			"timestampNano", tsNano,
		)
	}

	utils.DumpJSON(events, os.Stdout)

	return events
}

// enrichEvents fetches the full record for each event via GetRecordsByIds.
// Errors are logged rather than fatal: a sample id may not exist, or the token
// may lack the object's scope.
func enrichEvents(ctx context.Context, conn *zoho.Connector, events []common.SubscriptionEvent) {
	for _, evt := range events {
		objectName, err := evt.ObjectName()
		if err != nil {
			slog.Warn("skipping enrichment: no object name", "error", err)

			continue
		}

		recordID, err := evt.RecordId()
		if err != nil {
			slog.Warn("skipping enrichment: no record id", "error", err)

			continue
		}

		slog.Info("enriching via GetRecordsByIds..", "object", objectName, "recordId", recordID)

		rows, err := conn.GetRecordsByIds(ctx, objectName, []string{recordID}, nil, nil)
		if err != nil {
			slog.Warn("enrichment failed", "object", objectName, "recordId", recordID, "error", err)

			continue
		}

		utils.DumpJSON(rows, os.Stdout)
	}
}

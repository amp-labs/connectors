package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/providers/hubspot"
	connTest "github.com/amp-labs/connectors/test/hubspot"
	"github.com/amp-labs/connectors/test/utils"
)

const samplePropertyChange = `{
  "appId": 4210286,
  "eventId": 100,
  "subscriptionId": 2902227,
  "portalId": 44237313,
  "occurredAt": 1730750483646,
  "subscriptionType": "contact.propertyChange",
  "attemptNumber": 0,
  "objectId": 74999542704,
  "changeSource": "CRM",
  "propertyName": "message",
  "propertyValue": "sample-value"
}`

const sampleRecordContact = `{
  "success": true,
  "recordId": "74999542704",
  "data": {
    "company": "Personalis",
    "createdate": "2024-11-04T21:51:26.472Z",
    "email": "lestermertz@larson.org",
    "firstname": "Lucious",
    "hs_all_contact_vids": "74999542704",
    "hs_calculated_phone_number": "+17087038093",
    "hs_calculated_phone_number_country_code": "US",
    "hs_currently_enrolled_in_prospecting_agent": "false",
    "hs_email_domain": "larson.org",
    "hs_is_contact": "true",
    "hs_is_unworked": "true",
    "hs_lifecyclestage_lead_date": "2024-11-04T21:51:26.472Z",
    "hs_membership_has_accessed_private_content": "0",
    "hs_object_id": "74999542704",
    "hs_object_source": "INTEGRATION",
    "hs_object_source_id": "2317233",
    "hs_object_source_label": "INTEGRATION",
    "hs_pipeline": "contacts-lifecycle-pipeline",
    "hs_registered_member": "0",
    "hs_searchable_calculated_phone_number": "7087038093",
    "lastmodifieddate": "2024-11-04T21:51:26.472Z",
    "lastname": "Spencer",
    "lifecyclestage": "lead",
    "phone": "7087038093",
    "website": "https://www.directcutting-edge.net/granular/reintermediate/infomediaries"
  }
}`

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	// Get the Hubspot connector.
	conn := connTest.GetHubspotConnector(ctx)
	defer utils.Close(conn)

	propMsg := hubspot.WebhookMessage{}

	if err := json.Unmarshal([]byte(samplePropertyChange), &propMsg); err != nil {
		utils.Fail("error unmarshalling property change message", "error", err)
	}

	recordResult, err := conn.GetWebhookResultFromWebhookMessage(ctx, &propMsg)
	if err != nil {
		utils.Fail("error getting record from webhook message", "error", err)
	}

	utils.DumpJSON(recordResult, os.Stdout)
}

package salesforce

import (
	"errors"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// sampleOutboundMessage mirrors the SOAP envelope Salesforce POSTs to an
// outbound message endpoint, per
// https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_om_outboundmessaging_understanding.htm
const sampleOutboundMessage = `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <soapenv:Body>
    <notifications xmlns="http://soap.sforce.com/2005/09/outbound">
      <OrganizationId>00Dxx0000001gPFEAY</OrganizationId>
      <ActionId>04kxx00000000blAAA</ActionId>
      <SessionId xsi:nil="true"/>
      <EnterpriseUrl>https://yourInstance.salesforce.com/services/Soap/c/60.0/00Dxx0000001gPF</EnterpriseUrl>
      <PartnerUrl>https://yourInstance.salesforce.com/services/Soap/u/60.0/00Dxx0000001gPF</PartnerUrl>
      <Notification>
        <Id>04lxx000000CaSdAAK</Id>
        <sObject xsi:type="sf:Account" xmlns:sf="urn:sobject.enterprise.soap.sforce.com">
          <sf:Id>001xx000003DGb2AAG</sf:Id>
          <sf:CreatedDate>2026-09-01T10:00:00.000Z</sf:CreatedDate>
          <sf:LastModifiedDate>2026-09-01T10:00:00.000Z</sf:LastModifiedDate>
          <sf:Name>Acme</sf:Name>
        </sObject>
      </Notification>
      <Notification>
        <Id>04lxx000000CaSeAAK</Id>
        <sObject xsi:type="sf:Account" xmlns:sf="urn:sobject.enterprise.soap.sforce.com">
          <sf:Id>001xx000003DGb3AAG</sf:Id>
          <sf:CreatedDate>2026-08-20T08:30:00.000Z</sf:CreatedDate>
          <sf:LastModifiedDate>2026-09-01T11:15:00.000Z</sf:LastModifiedDate>
          <sf:Name xsi:nil="true"/>
        </sObject>
      </Notification>
    </notifications>
  </soapenv:Body>
</soapenv:Envelope>`

func parseSampleEvents(t *testing.T) []common.SubscriptionEvent {
	t.Helper()

	envelope, err := ParseOutboundMessage([]byte(sampleOutboundMessage))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	events, err := envelope.SubscriptionEventList()
	if err != nil {
		t.Fatalf("unexpected expand error: %v", err)
	}

	return events
}

func TestParseOutboundMessageExpandsNotifications(t *testing.T) {
	t.Parallel()

	events := parseSampleEvents(t)

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	for index, expectedID := range []string{"001xx000003DGb2AAG", "001xx000003DGb3AAG"} {
		recordID, err := events[index].RecordId()
		if err != nil {
			t.Fatalf("RecordId(%d): %v", index, err)
		}

		if recordID != expectedID {
			t.Errorf("RecordId(%d) = %q, want %q", index, recordID, expectedID)
		}

		objectName, err := events[index].ObjectName()
		if err != nil {
			t.Fatalf("ObjectName(%d): %v", index, err)
		}

		if objectName != "Account" {
			t.Errorf("ObjectName(%d) = %q, want Account", index, objectName)
		}
	}
}

func TestOutboundMessageEventTypeInference(t *testing.T) {
	t.Parallel()

	events := parseSampleEvents(t)

	// First notification: CreatedDate == LastModifiedDate → the save that
	// fired the flow created the record.
	eventType, err := events[0].EventType()
	if err != nil {
		t.Fatalf("EventType(0): %v", err)
	}

	if eventType != common.SubscriptionEventTypeCreate {
		t.Errorf("EventType(0) = %q, want create", eventType)
	}

	rawName, err := events[0].RawEventName()
	if err != nil || rawName != "CREATE" {
		t.Errorf("RawEventName(0) = %q, %v; want CREATE", rawName, err)
	}

	// Second notification: LastModifiedDate after CreatedDate → update.
	eventType, err = events[1].EventType()
	if err != nil {
		t.Fatalf("EventType(1): %v", err)
	}

	if eventType != common.SubscriptionEventTypeUpdate {
		t.Errorf("EventType(1) = %q, want update", eventType)
	}
}

func TestOutboundMessageEventTimestampAndNilFields(t *testing.T) {
	t.Parallel()

	events := parseSampleEvents(t)

	nano, err := events[1].EventTimeStampNano()
	if err != nil {
		t.Fatalf("EventTimeStampNano: %v", err)
	}

	// 2026-09-01T11:15:00Z
	const expectedNano = int64(1788261300000000000)
	if nano != expectedNano {
		t.Errorf("EventTimeStampNano = %d, want %d", nano, expectedNano)
	}

	// xsi:nil fields parse as explicit nils rather than empty strings.
	raw, err := events[1].RawMap()
	if err != nil {
		t.Fatalf("RawMap: %v", err)
	}

	fields, ok := raw[omKeySObject].(map[string]any)
	if !ok {
		t.Fatalf("sObject missing from raw map: %v", raw)
	}

	if value, exists := fields["Name"]; !exists || value != nil {
		t.Errorf("nil Name field = %v (exists=%t), want nil", value, exists)
	}

	if raw[omKeyOrganizationID] != "00Dxx0000001gPFEAY" {
		t.Errorf("organizationId = %v", raw[omKeyOrganizationID])
	}
}

func TestParseOutboundMessageRejectsNonSoapBodies(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"JSON":             `{"not": "soap"}`,
		"empty":            ``,
		"no notifications": `<?xml version="1.0"?><Envelope><Body></Body></Envelope>`,
		"other SOAP":       `<?xml version="1.0"?><Envelope><Body><other/></Body></Envelope>`, //nolint:dupword
	} {
		if _, err := ParseOutboundMessage([]byte(body)); !errors.Is(err, errNotOutboundMessage) {
			t.Errorf("%s: expected errNotOutboundMessage, got %v", name, err)
		}
	}
}

func TestBuildOutboundMessageAck(t *testing.T) {
	t.Parallel()

	ack := string(BuildOutboundMessageAck(true))
	if !strings.Contains(ack, "<Ack>true</Ack>") ||
		!strings.Contains(ack, `<notificationsResponse xmlns="http://soap.sforce.com/2005/09/outbound">`) {
		t.Errorf("ack body malformed:\n%s", ack)
	}

	nack := string(BuildOutboundMessageAck(false))
	if !strings.Contains(nack, "<Ack>false</Ack>") {
		t.Errorf("nack body malformed:\n%s", nack)
	}
}

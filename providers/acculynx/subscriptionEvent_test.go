package acculynx

import (
	"errors"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"gotest.tools/v3/assert"
)

func TestSubscriptionEvent_ObjectName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		topic    string
		expected string
		wantErr  error
	}{
		{"contact_added", objectContacts, nil},
		{"contact_changed", objectContacts, nil},
		{"contact.custom-field.value_changed", objectContacts, nil},
		{"contact.custom-field.status_changed", objectContacts, nil},
		{"job_created", objectJobs, nil},
		{"job_updated", objectJobs, nil},
		{"job.milestone.current_changed", objectJobs, nil},
		{"job.financials.approved-value_changed", objectJobs, nil},
		{"unsupported_topic", "", errUnsupportedTopicName},
	}

	for _, tc := range cases {
		t.Run(tc.topic, func(t *testing.T) {
			t.Parallel()

			evt := SubscriptionEvent{eventFieldTopicName: tc.topic}

			got, err := evt.ObjectName()
			if tc.wantErr != nil {
				assert.Assert(t, errors.Is(err, tc.wantErr), "expected %v, got %v", tc.wantErr, err)

				return
			}

			assert.NilError(t, err)
			assert.Equal(t, got, tc.expected)
		})
	}
}

func TestSubscriptionEvent_EventType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		topic    string
		expected common.SubscriptionEventType
	}{
		{"contact_added", common.SubscriptionEventTypeCreate},
		{"job_created", common.SubscriptionEventTypeCreate},
		{"job.appointments.initial_created", common.SubscriptionEventTypeCreate},
		{"contact_changed", common.SubscriptionEventTypeUpdate},
		{"job_updated", common.SubscriptionEventTypeUpdate},
		{"job.milestone.current_changed", common.SubscriptionEventTypeUpdate},
		{"job.invoice_voided", common.SubscriptionEventTypeUpdate},
		{"job.something_weird", common.SubscriptionEventTypeOther},
	}

	for _, tc := range cases {
		t.Run(tc.topic, func(t *testing.T) {
			t.Parallel()

			evt := SubscriptionEvent{eventFieldTopicName: tc.topic}

			got, err := evt.EventType()

			assert.NilError(t, err)
			assert.Equal(t, got, tc.expected)
		})
	}
}

func TestSubscriptionEvent_Workspace_ReturnsSubscriptionId(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		eventFieldTopicName:      "job_created",
		eventFieldSubscriptionID: "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8",
		eventFieldEvent: map[string]any{
			objectWrapperJob: map[string]any{
				innerFieldID: "185548c3-3de4-4dca-b79d-e8cf38c0f776",
			},
		},
	}

	got, err := evt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, got, "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8")
}

func TestSubscriptionEvent_Workspace_MissingSubscriptionIdReturnsError(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{eventFieldTopicName: "job_created"}

	_, err := evt.Workspace()
	assert.Assert(t, errors.Is(err, errMissingSubscriptionID))
}

func TestSubscriptionEvent_RecordId_RoutesByObjectType(t *testing.T) {
	t.Parallel()

	jobEvt := SubscriptionEvent{
		eventFieldTopicName: "job.milestone.current_changed",
		eventFieldEvent: map[string]any{
			objectWrapperJob: map[string]any{
				innerFieldID: "job-123",
			},
		},
	}

	jobID, err := jobEvt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, jobID, "job-123")

	contactEvt := SubscriptionEvent{
		eventFieldTopicName: "contact_changed",
		eventFieldEvent: map[string]any{
			objectWrapperContact: map[string]any{
				innerFieldID: "contact-456",
			},
		},
	}

	contactID, err := contactEvt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, contactID, "contact-456")
}

func TestSubscriptionEvent_RecordId_MissingObjectWrapperReturnsError(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		eventFieldTopicName: "job_created",
		eventFieldEvent:     map[string]any{},
	}

	_, err := evt.RecordId()
	assert.Assert(t, errors.Is(err, errMissingObjectWrapper))
}

func TestSubscriptionEvent_RecordId_MissingIdReturnsError(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		eventFieldTopicName: "job_created",
		eventFieldEvent: map[string]any{
			objectWrapperJob: map[string]any{},
		},
	}

	_, err := evt.RecordId()
	assert.Assert(t, errors.Is(err, errMissingRecordID))
}

func TestSubscriptionEvent_EventTimeStampNano_ParsesRFC3339Nano(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		eventFieldEventDateTime: "2023-07-11T22:38:57.1993216Z",
	}

	got, err := evt.EventTimeStampNano()
	assert.NilError(t, err)

	expected := time.Date(2023, 7, 11, 22, 38, 57, 199321600, time.UTC).UnixNano()
	assert.Equal(t, got, expected)
}

func TestSubscriptionEvent_EventTimeStampNano_RejectsBogus(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{eventFieldEventDateTime: "not a timestamp"}

	_, err := evt.EventTimeStampNano()
	assert.Assert(t, errors.Is(err, errUnparsableEventTime))
}

func TestSubscriptionEvent_RawEventName_ReturnsVerbatim(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{eventFieldTopicName: "job.milestone.current_changed"}

	got, err := evt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, got, "job.milestone.current_changed")
}

func TestSubscriptionEvent_RawEventName_MissingReturnsError(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{}

	_, err := evt.RawEventName()
	assert.Assert(t, errors.Is(err, errMissingTopicName))
}

func TestSubscriptionEvent_RawMap_ReturnsCloneNotAlias(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		eventFieldTopicName: "job_created",
		eventFieldEvent: map[string]any{
			objectWrapperJob: map[string]any{innerFieldID: "job-1"},
		},
	}

	clone, err := evt.RawMap()
	assert.NilError(t, err)

	// Mutating the clone should not affect the original.
	clone[eventFieldTopicName] = "tampered"
	got, _ := evt.RawEventName()
	assert.Equal(t, got, "job_created", "original should be unchanged")
}

func TestSubscriptionEvent_UpdatedFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		topic    string
		expected []string
	}{
		// Specific-change topics — each maps to a single field name verified
		// against live AccuLynx test-event payloads.
		{"job.milestone.current_changed", []string{"milestone"}},
		{"job.milestone.status.current_changed", []string{"milestone"}},
		{"job.financials.approved-value_changed", []string{"financials"}},
		{"job.category_changed", []string{"jobCategory"}},
		{"job.work-type_changed", []string{"workType"}},
		{"job.trade-type_changed", []string{"tradeTypes"}},
		{"job.contacts.primary_changed", []string{"contacts"}},
		{"job.representatives.company_assigned", []string{"companyRepresentative"}},
		{"job.representatives.company_changed", []string{"companyRepresentative"}},
		{"job.appointments.initial_created", []string{"initialAppointment"}},
		{"job.appointments.initial_updated", []string{"initialAppointment"}},
		{"job.invoice_updated", []string{"invoice"}},
		{"job.invoice_voided", []string{"invoice"}},
		{"job.custom-field.value_changed", []string{"customField"}},
		{"job.custom-field.status_changed", []string{"customField"}},
		{"contact.custom-field.value_changed", []string{"customField"}},
		{"contact.custom-field.status_changed", []string{"customField"}},
		{"job.accounting.integration-status.current_changed", []string{"accounting"}},
		// Generic update topics — field is unknown, empty slice expected.
		{"job_updated", []string{}},
		{"contact_changed", []string{}},
		// Create topics — not field-specific, empty slice expected.
		{"job_created", []string{}},
		{"contact_added", []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.topic, func(t *testing.T) {
			t.Parallel()

			evt := SubscriptionEvent{eventFieldTopicName: tc.topic}

			got, err := evt.UpdatedFields()
			assert.NilError(t, err)
			assert.DeepEqual(t, got, tc.expected)
		})
	}
}

func TestSubscriptionEvent_UpdatedFields_MissingTopicReturnsError(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{}

	_, err := evt.UpdatedFields()
	assert.Assert(t, errors.Is(err, errMissingTopicName))
}

// Real production payload captured live for contact_added — round-trips through
// every accessor to lock in the verified wire format.
func TestSubscriptionEvent_RealPayload_ContactAdded(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"topicName":      "contact_added",
		"eventDateTime":  "2026-05-26T14:23:55.4999782Z",
		"eventId":        "38e4c045-2a6c-43f2-8309-ac8b5fc3fc2b",
		"subscriptionId": "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8",
		"event": map[string]any{
			"contact": map[string]any{
				"id":    "eadaaa11-1276-4166-bb93-db02f46b39a2",
				"date":  "2026-05-26T14:23:55.2976156Z",
				"_link": "https://api.acculynx.com/api/v2/contacts/eadaaa11-1276-4166-bb93-db02f46b39a2",
			},
		},
	}

	obj, err := evt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, obj, objectContacts)

	evtType, err := evt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, evtType, common.SubscriptionEventTypeCreate)

	ws, err := evt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, ws, "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8")

	rid, err := evt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, rid, "eadaaa11-1276-4166-bb93-db02f46b39a2")

	raw, err := evt.RawEventName()
	assert.NilError(t, err)
	assert.Equal(t, raw, "contact_added")

	ts, err := evt.EventTimeStampNano()
	assert.NilError(t, err)
	assert.Assert(t, ts > 0)

	fields, err := evt.UpdatedFields()
	assert.NilError(t, err)
	assert.DeepEqual(t, fields, []string{})
}

// Real production payload captured live for job.milestone.current_changed.
func TestSubscriptionEvent_RealPayload_JobMilestoneChanged(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"topicName":      "job.milestone.current_changed",
		"eventDateTime":  "2026-05-26T13:28:42.400449Z",
		"eventId":        "8ee2d58a-3062-462e-b1c2-57a9dd24921d",
		"subscriptionId": "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8",
		"event": map[string]any{
			"job": map[string]any{
				"id": "0fb0dfde-6cff-4590-a864-9823e10e58c0",
				"milestone": map[string]any{
					"id":        "5c60878e-8fb1-42a9-883d-9b734e800e7e",
					"name":      "MilestoneName439637d0-d451-4a7a-a845-fdc6b3ed408e",
					"isCurrent": true,
				},
				"_link": "https://api.acculynx.com/api/v2/jobs/0fb0dfde-6cff-4590-a864-9823e10e58c0",
			},
		},
	}

	obj, err := evt.ObjectName()
	assert.NilError(t, err)
	assert.Equal(t, obj, objectJobs)

	evtType, err := evt.EventType()
	assert.NilError(t, err)
	assert.Equal(t, evtType, common.SubscriptionEventTypeUpdate)

	ws, err := evt.Workspace()
	assert.NilError(t, err)
	assert.Equal(t, ws, "6541d9e1-12c1-45b8-b5bd-5ffa8849a4b8")

	rid, err := evt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, rid, "0fb0dfde-6cff-4590-a864-9823e10e58c0")

	fields, err := evt.UpdatedFields()
	assert.NilError(t, err)
	assert.DeepEqual(t, fields, []string{"milestone"})
}

// Real production payload captured live for job.invoice_updated — verifies the
// invoice (singular) field path our mapping returns.
func TestSubscriptionEvent_RealPayload_JobInvoiceUpdated(t *testing.T) {
	t.Parallel()

	evt := SubscriptionEvent{
		"topicName":      "job.invoice_updated",
		"eventDateTime":  "2026-05-27T06:59:49.8688855Z",
		"eventId":        "6f8adfd1-263a-466e-9031-bb005610e2ff",
		"subscriptionId": "404d797c-af65-4ca3-a38e-a490ca222ad2",
		"event": map[string]any{
			"job": map[string]any{
				"id": "9e22b9e5-787f-4052-9f47-1aad34f6f515",
				"invoice": map[string]any{
					"id":              "39e2c252-eb64-433b-a0c8-02258241df93",
					"costTotal":       float64(23132),
					"priceTotal":      float64(15358),
					"amountCollected": float64(27235),
				},
				"_link": "https://api.acculynx.com/api/v2/jobs/9e22b9e5-787f-4052-9f47-1aad34f6f515",
			},
		},
	}

	rid, err := evt.RecordId()
	assert.NilError(t, err)
	assert.Equal(t, rid, "9e22b9e5-787f-4052-9f47-1aad34f6f515")

	fields, err := evt.UpdatedFields()
	assert.NilError(t, err)
	assert.DeepEqual(t, fields, []string{"invoice"})
}

// Verifies the special-case handling for *.custom-field.status_changed payloads:
// AccuLynx omits the parent contact/job id entirely (only the changed
// custom field's own id is delivered), so RecordId must return
// errParentRecordIDUnavailable rather than a misleading id.
func TestSubscriptionEvent_RecordId_CustomFieldStatusChanged_NoParentID(t *testing.T) {
	t.Parallel()

	cases := []string{
		"contact.custom-field.status_changed",
		"job.custom-field.status_changed",
	}

	for _, topic := range cases {
		t.Run(topic, func(t *testing.T) {
			t.Parallel()

			// Real wire shape: no contact/job wrapper — only customField.
			evt := SubscriptionEvent{
				eventFieldTopicName:      topic,
				eventFieldSubscriptionID: "sub-xyz",
				eventFieldEvent: map[string]any{
					"customField": map[string]any{
						"id":     "cf-id",
						"label":  "Some Label",
						"status": "Unknown",
					},
				},
			}

			// Other accessors still work.
			obj, err := evt.ObjectName()
			assert.NilError(t, err)
			assert.Assert(t, obj == objectContacts || obj == objectJobs)

			ws, err := evt.Workspace()
			assert.NilError(t, err)
			assert.Equal(t, ws, "sub-xyz")

			// But RecordId fails with the dedicated sentinel so consumers can skip enrichment.
			_, err = evt.RecordId()
			assert.Assert(t, errors.Is(err, errParentRecordIDUnavailable))
		})
	}
}

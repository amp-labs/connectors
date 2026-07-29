package acculynx

import (
	"slices"

	"github.com/amp-labs/connectors/common"
)

// appointmentJobAssociation and appointmentUserAssociation are the association
// keys Hatch's config requests for calendars/appointments (see the server's
// getHatchAssociations, which returns ["jobs", "users"]). They match the target
// object names the server hydrates.
const (
	appointmentJobAssociation  = objectJobs
	appointmentUserAssociation = objectUsers
)

// attachAppointmentAssociations attaches the Appointment->Job and
// Appointment->User edges to appointment rows produced by the calendars/
// appointments fan-out. Both are reference-shape associations (ObjectId only,
// the server hydrates via GetRecordsByIds) — the same shape as outreach's
// collectAssociations, and unlike the embedded Job<->Contact edge above.
//
//   - Appointment->Job: the job id rides on the appointment body (jobId). It is
//     empty for non-job calendar events (e.g. Personal), so the edge is emitted
//     only when present.
//   - Appointment->User: the appointment body carries no calendar/user id — the
//     calendarId is the fan-out path parent, so it is threaded in. A calendar's
//     id equals its user's id in AccuLynx; company/crew calendars are not users
//     and are dropped downstream when hydration 404s (see GetRecordsByIds).
//
// It is a no-op for any object other than calendars/appointments, so the nested
// fan-out can call it unconditionally.
func attachAppointmentAssociations(params common.ReadParams, calendarID string, rows []common.ReadResultRow) {
	if params.ObjectName != "calendars/appointments" {
		return
	}

	wantJob := slices.Contains(params.AssociatedObjects, appointmentJobAssociation)
	wantUser := slices.Contains(params.AssociatedObjects, appointmentUserAssociation)

	if !wantJob && !wantUser {
		return
	}

	for idx := range rows {
		if wantJob {
			if jobID, _ := rows[idx].Raw["jobId"].(string); jobID != "" {
				addAssociation(&rows[idx], appointmentJobAssociation, jobID)
			}
		}

		if wantUser && calendarID != "" {
			addAssociation(&rows[idx], appointmentUserAssociation, calendarID)
		}
	}
}

// addAssociation appends a reference-shape association (ObjectId only) under the
// given key, initialising the row's association map on first use.
func addAssociation(row *common.ReadResultRow, key, objectID string) {
	if row.Associations == nil {
		row.Associations = make(map[string][]common.Association)
	}

	row.Associations[key] = append(row.Associations[key], common.Association{ObjectId: objectID})
}

// jobContactsAssociation is the association key Hatch's config requests (see the
// server's getHatchAssociations). It matches the embedded "contacts" array that
// AccuLynx returns on /jobs when the read is issued with ?includes=contacts.
const jobContactsAssociation = "contacts"

// extractJobContacts attaches each job's embedded contacts as associations.
//
// AccuLynx returns the contacts inline on the job payload (with ?includes=contacts)
// as an array of { contact: { id, ... }, isPrimary } entries. Unlike the HousecallPro
// job->customer association (a single embedded object), a job carries an array of
// contacts, so we emit one common.Association per entry, populating Raw from the
// embedded contact (embed path -> server attaches it directly, no extra fetch) and
// carrying isPrimary through in ProviderAssociationMetadata when present.
func extractJobContacts(rows []common.ReadResultRow) {
	for idx := range rows {
		assocs := jobContactAssociations(rows[idx].Raw)
		if len(assocs) == 0 {
			continue
		}

		if rows[idx].Associations == nil {
			rows[idx].Associations = make(map[string][]common.Association)
		}

		rows[idx].Associations[jobContactsAssociation] = assocs
	}
}

// jobContactAssociations builds the contact associations from a job's raw payload.
func jobContactAssociations(raw map[string]any) []common.Association {
	entries, ok := raw["contacts"].([]any)
	if !ok || len(entries) == 0 {
		return nil
	}

	assocs := make([]common.Association, 0, len(entries))

	for _, entry := range entries {
		if assoc, ok := jobContactAssociation(entry); ok {
			assocs = append(assocs, assoc)
		}
	}

	return assocs
}

// jobContactAssociation converts a single job-contact entry into an Association.
// Returns false when the entry is malformed or carries no contact id.
func jobContactAssociation(entry any) (common.Association, bool) {
	link, ok := entry.(map[string]any)
	if !ok {
		return common.Association{}, false
	}

	contact, ok := link["contact"].(map[string]any)
	if !ok {
		return common.Association{}, false
	}

	id, _ := contact["id"].(string)
	if id == "" {
		return common.Association{}, false
	}

	assoc := common.Association{ObjectId: id, Raw: contact}

	if v, ok := link["isPrimary"]; ok {
		assoc.ProviderAssociationMetadata = map[string]any{"isPrimary": v}
	}

	return assoc, true
}

package acculynx

import "github.com/amp-labs/connectors/common"

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

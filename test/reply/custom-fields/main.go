// Live scenario: custom fields on contacts and contact-accounts.
//
// It first reports the custom fields Reply already holds, so fields created by
// hand in the Reply UI show up. Then, with throwaway custom fields, it checks that
//   - metadata exposes a custom field natively (IsCustom) under its title,
//   - a value written by title can be read back by that title,
//   - a ReadResult round-trips back into a write, and
//   - a custom field titled like a built-in field (contact notes, account name)
//     never replaces the built-in.
//
// Account custom fields are a feature Reply gates per team: when the team lacks
// it the account half is skipped. Everything created is deleted at the end,
// records first because Reply refuses to delete a field that holds values.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	replyconn "github.com/amp-labs/connectors/providers/reply"
	"github.com/amp-labs/connectors/test/reply"
	"github.com/amp-labs/connectors/test/utils"
)

const (
	// maxReadPages bounds the search for a record, the sandbox is small.
	maxReadPages = 10

	// customFieldsProperty is where Reply puts the values on a record.
	customFieldsProperty = "customFields"
)

type created struct {
	objectName string
	recordID   string
}

type scenario struct {
	conn *replyconn.Connector
	// Deleted in reverse order of creation.
	created []created
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	s := &scenario{conn: reply.GetReplyConnector(ctx)}

	suffix := fmt.Sprintf("%d", time.Now().Unix())

	s.existingFieldsScenario(ctx)

	s.contactScenario(ctx, suffix)
	s.contactCollisionScenario(ctx, suffix)

	if s.accountScenario(ctx, suffix) {
		s.accountCollisionScenario(ctx, suffix)
	}

	slog.Info("=== Cleanup: delete everything the scenario created, newest first ===")
	s.cleanup(ctx)
	slog.Info("Custom fields scenario passed")
}

// existingLabelField is the built-in field printed next to each record, so a
// row can be recognised in the Reply UI.
//
//nolint:gochecknoglobals
var existingLabelField = map[string]string{
	"contacts":         "email",
	"contact-accounts": "name",
}

// existingFieldsScenario reports the custom fields Reply already holds before
// the scenario creates anything: every custom field the connector exposes per
// object, and the records of the first pages that hold a value for any of them.
// Fields created by hand in the Reply UI therefore show up here.
func (s *scenario) existingFieldsScenario(ctx context.Context) {
	slog.Info("=== Existing custom fields: what Reply already holds, including fields created in the UI ===")

	for _, objectName := range []string{"contacts", "contact-accounts"} {
		s.reportExistingFields(ctx, objectName)
	}
}

func (s *scenario) reportExistingFields(ctx context.Context, objectName string) {
	metadata, err := s.conn.ListObjectMetadata(ctx, []string{objectName})
	if err != nil {
		s.abort(ctx, "list "+objectName+" metadata", err)
	}

	objectMetadata := metadata.GetObjectMetadata(objectName)
	if objectMetadata == nil {
		s.abort(ctx, "no metadata for "+objectName, errUnexpected)
	}

	custom := make(map[string]common.FieldMetadata)
	titles := make([]string, 0)

	for name, field := range objectMetadata.Fields {
		if field.IsCustom != nil && *field.IsCustom {
			custom[name] = field
			titles = append(titles, name)
		}
	}

	slices.Sort(titles)

	slog.Info("Custom fields exposed by metadata",
		"object", objectName, "count", len(titles), "titles", titles)
	utils.DumpJSON(custom, os.Stdout)

	if len(titles) == 0 {
		return
	}

	s.reportRecordsWithValues(ctx, objectName, titles)
}

// reportRecordsWithValues reads the first pages and prints every record that
// holds a custom value, exactly as the connector returns it: Fields carries the
// values resolved by title and Raw carries what Reply sent, customFields included.
func (s *scenario) reportRecordsWithValues(ctx context.Context, objectName string, titles []string) {
	label := existingLabelField[objectName]
	params := common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(append([]string{label}, titles...)...),
	}

	scanned, hits := 0, 0

	for range maxReadPages {
		res, err := s.conn.Read(ctx, params)
		if err != nil {
			s.abort(ctx, "read "+objectName, err)
		}

		for _, row := range res.Data {
			scanned++

			if !holdsCustomValue(row, titles) {
				continue
			}

			hits++

			slog.Info("Record holds custom field values", "object", objectName,
				"recordId", row.Id, label, row.Fields[label])
			utils.DumpJSON(row, os.Stdout)
		}

		if res.Done {
			break
		}

		params.NextPage = res.NextPage
	}

	slog.Info("Scanned existing records", "object", objectName, "scanned", scanned, "withCustomValues", hits)
}

// holdsCustomValue reports whether the record carries a value for any custom
// field, resolved into Fields or present in Raw.customFields.
func holdsCustomValue(row common.ReadResultRow, titles []string) bool {
	for _, title := range titles {
		if row.Fields[strings.ToLower(title)] != nil {
			return true
		}
	}

	raw, _ := row.Raw[customFieldsProperty].([]any)

	return len(raw) > 0
}

func (s *scenario) contactScenario(ctx context.Context, suffix string) {
	slog.Info("=== Contacts: create a custom field, write it by title, read it back, round-trip a ReadResult ===")

	title := "Live Contact CF " + suffix

	s.write(ctx, "create contact custom field", common.WriteParams{
		ObjectName: "custom-fields",
		RecordData: map[string]any{"title": title, "fieldType": "text"},
	})

	s.expectCustomMetadata(ctx, "contacts", title, common.ValueTypeString)

	contact := s.write(ctx, "create contact with value named by title", common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"email":     fmt.Sprintf("live.customfields.%s@example.com", suffix),
			"firstName": "Live",
			"lastName":  "CustomFields",
			title:       "one",
		},
	})

	s.expectValue(ctx, "contacts", contact.RecordId, title, "one")

	// Lowercased key, as ReadResult.Fields would carry it.
	s.write(ctx, "update contact with lowercased title", common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contact.RecordId,
		RecordData: map[string]any{strings.ToLower(title): "two"},
	})

	s.expectValue(ctx, "contacts", contact.RecordId, title, "two")

	// Full ReadResult round-trip: the customFields entries (key form) and the
	// title-named field both come back, and the title-named value is edited.
	row := s.findRow(ctx, "contacts", contact.RecordId, "email", "firstName", "customFields", title)
	row.Fields[strings.ToLower(title)] = "three"

	s.write(ctx, "update contact from a ReadResult", common.WriteParams{
		ObjectName: "contacts",
		RecordId:   contact.RecordId,
		RecordData: row.Fields,
	})

	s.expectValue(ctx, "contacts", contact.RecordId, title, "three")
}

// accountScenario reports whether the team has account custom fields.
func (s *scenario) accountScenario(ctx context.Context, suffix string) bool {
	slog.Info("=== Contact accounts: same flow with a number field (skipped if the team lacks the feature) ===")

	title := "Live Account CF " + suffix

	definitionParams := common.WriteParams{
		ObjectName: "custom-fields",
		RecordData: map[string]any{"title": title, "fieldType": "number", "entityType": "account"},
	}

	slog.Info("Write request", "operation", "create account custom field", "object", "custom-fields")
	utils.DumpJSON(definitionParams.RecordData, os.Stdout)

	definition, err := s.conn.Write(ctx, definitionParams)
	if err != nil {
		if errors.Is(err, common.ErrForbidden) {
			slog.Warn("Account custom fields are not available for this team, skipping the account half",
				"error", err)

			return false
		}

		s.abort(ctx, "create account custom field", err)
	}

	s.created = append(s.created, created{"custom-fields", definition.RecordId})
	slog.Info("Write result", "operation", "create account custom field", "recordId", definition.RecordId)
	utils.DumpJSON(definition, os.Stdout)

	s.expectCustomMetadata(ctx, "contact-accounts", title, common.ValueTypeFloat)

	account := s.write(ctx, "create account with a number value named by title", common.WriteParams{
		ObjectName: "contact-accounts",
		RecordData: map[string]any{"name": "Live CustomFields " + suffix, title: float64(7)},
	})

	s.expectValue(ctx, "contact-accounts", account.RecordId, title, "7")

	// PUT replaces the account, so the name and the owner go along.
	row := s.findRow(ctx, "contact-accounts", account.RecordId, "name", "ownerUserId", "customFields", title)
	row.Fields[strings.ToLower(title)] = "9"

	s.write(ctx, "update account from a ReadResult", common.WriteParams{
		ObjectName: "contact-accounts",
		RecordId:   account.RecordId,
		RecordData: row.Fields,
	})

	s.expectValue(ctx, "contact-accounts", account.RecordId, title, "9")

	return true
}

// contactCollisionScenario creates a custom field titled exactly like the
// built-in "notes" and checks that the built-in value is left alone.
func (s *scenario) contactCollisionScenario(ctx context.Context, suffix string) {
	slog.Info("=== Contacts collision: a custom field titled like the built-in notes must not replace it ===")

	field := s.write(ctx, "create custom field titled Notes", common.WriteParams{
		ObjectName: "custom-fields",
		RecordData: map[string]any{"title": "Notes", "fieldType": "text"},
	})

	contact := s.write(ctx, "create contact with built-in notes and a custom Notes value", common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"email":     fmt.Sprintf("live.collision.%s@example.com", suffix),
			"firstName": "Live",
			"notes":     "built-in note",
			// By id: create takes the key form, which the connector derives.
			"customFields": []any{map[string]any{"id": s.fieldID(ctx, field), "value": "custom note"}},
		},
	})

	s.expectBuiltIn(ctx, "contacts", contact.RecordId, "notes", "built-in note", "Notes")
}

// accountCollisionScenario does the same for the built-in account "name".
func (s *scenario) accountCollisionScenario(ctx context.Context, suffix string) {
	slog.Info("=== Contact accounts collision: a custom field titled like the built-in name must not replace it ===")

	field := s.write(ctx, "create account custom field titled Name", common.WriteParams{
		ObjectName: "custom-fields",
		RecordData: map[string]any{"title": "Name", "fieldType": "text", "entityType": "account"},
	})

	name := "Live Collision " + suffix

	account := s.write(ctx, "create account with a built-in name and a custom Name value", common.WriteParams{
		ObjectName: "contact-accounts",
		RecordData: map[string]any{
			"name":         name,
			"customFields": []any{map[string]any{"id": s.fieldID(ctx, field), "value": "custom name"}},
		},
	})

	s.expectBuiltIn(ctx, "contact-accounts", account.RecordId, "name", name, "Name")
}

// expectBuiltIn checks that the built-in field kept its value and that the
// colliding custom field is not exposed as a native field either.
func (s *scenario) expectBuiltIn(
	ctx context.Context, objectName, recordID, builtIn string, expected any, customTitle string,
) {
	row := s.findRow(ctx, objectName, recordID, builtIn)
	if row.Fields[builtIn] != expected {
		s.abort(ctx, "built-in "+builtIn+" was replaced in "+objectName,
			fmt.Errorf("%w: want %v, got %v", errUnexpected, expected, row.Fields[builtIn]))
	}

	metadata, err := s.conn.ListObjectMetadata(ctx, []string{objectName})
	if err != nil {
		s.abort(ctx, "list "+objectName+" metadata", err)
	}

	if objectMetadata := metadata.GetObjectMetadata(objectName); objectMetadata != nil {
		if _, exposed := objectMetadata.Fields[customTitle]; exposed {
			s.abort(ctx, "colliding custom field was exposed in "+objectName, errUnexpected)
		}
	}

	slog.Info("Built-in field kept its value and the colliding custom field is not exposed",
		"object", objectName, "builtIn", builtIn, "customTitle", customTitle)
	utils.DumpJSON(row, os.Stdout)
}

func (s *scenario) fieldID(ctx context.Context, field *common.WriteResult) int {
	id, err := strconv.Atoi(field.RecordId)
	if err != nil {
		s.abort(ctx, "custom field id is not numeric", err)
	}

	return id
}

var errUnexpected = errors.New("unexpected live result")

func (s *scenario) expectCustomMetadata(
	ctx context.Context, objectName, title string, valueType common.ValueType,
) {
	metadata, err := s.conn.ListObjectMetadata(ctx, []string{objectName})
	if err != nil {
		s.abort(ctx, "list "+objectName+" metadata", err)
	}

	objectMetadata := metadata.GetObjectMetadata(objectName)
	if objectMetadata == nil {
		s.abort(ctx, "no metadata for "+objectName, errUnexpected)
	}

	field, found := objectMetadata.Fields[title]
	if !found || field.ValueType != valueType || field.IsCustom == nil || !*field.IsCustom {
		s.abort(ctx, "custom field missing from "+objectName+" metadata",
			fmt.Errorf("%w: %+v", errUnexpected, field))
	}

	slog.Info("Metadata exposes custom field natively, flagged IsCustom",
		"object", objectName, "field", title, "totalFields", len(objectMetadata.Fields))
	utils.DumpJSON(map[string]any{title: field}, os.Stdout)
}

func (s *scenario) expectValue(ctx context.Context, objectName, recordID, title, expected string) {
	row := s.findRow(ctx, objectName, recordID, title)

	if got := row.Fields[strings.ToLower(title)]; got != expected {
		s.abort(ctx, "unexpected custom field value in "+objectName,
			fmt.Errorf("%w: want %q, got %v", errUnexpected, expected, got))
	}

	slog.Info("Custom field value verified in ReadResult.Fields (Raw is untouched)",
		"object", objectName, "recordId", recordID, "field", strings.ToLower(title), "value", expected)
	utils.DumpJSON(row, os.Stdout)
}

// findRow reads the object until the record turns up.
func (s *scenario) findRow(
	ctx context.Context, objectName, recordID string, fields ...string,
) common.ReadResultRow {
	params := common.ReadParams{ObjectName: objectName, Fields: connectors.Fields(fields...)}

	for range maxReadPages {
		res, err := s.conn.Read(ctx, params)
		if err != nil {
			s.abort(ctx, "read "+objectName, err)
		}

		for _, row := range res.Data {
			if row.Id == recordID {
				return row
			}
		}

		if res.Done {
			break
		}

		params.NextPage = res.NextPage
	}

	s.abort(ctx, "record not found in "+objectName, fmt.Errorf("%w: id %s", errUnexpected, recordID))

	return common.ReadResultRow{}
}

func (s *scenario) write(ctx context.Context, name string, params common.WriteParams) *common.WriteResult {
	slog.Info("Write request", "operation", name, "object", params.ObjectName, "recordId", params.RecordId)
	utils.DumpJSON(params.RecordData, os.Stdout)

	res, err := s.conn.Write(ctx, params)
	if err != nil {
		s.abort(ctx, name, err)
	}

	if params.RecordId == "" {
		s.created = append(s.created, created{params.ObjectName, res.RecordId})
	}

	slog.Info("Write result", "operation", name, "object", params.ObjectName, "recordId", res.RecordId)
	utils.DumpJSON(res, os.Stdout)

	return res
}

// cleanup deletes what the scenario created, newest first. A failure is logged
// rather than fatal so that the remaining records are still attempted.
func (s *scenario) cleanup(ctx context.Context) {
	// An interrupt cancels ctx, which must not stop the cleanup it triggers.
	ctx = context.WithoutCancel(ctx)

	for index := len(s.created) - 1; index >= 0; index-- {
		item := s.created[index]

		deleted, err := s.conn.Delete(ctx, common.DeleteParams{
			ObjectName: item.objectName,
			RecordId:   item.recordID,
		})
		if err != nil {
			slog.Warn("Cleanup failed, delete it manually", "object", item.objectName,
				"recordId", item.recordID, "error", err)

			continue
		}

		slog.Info("Deleted", "object", item.objectName, "recordId", item.recordID)
		utils.DumpJSON(deleted, os.Stdout)
	}
}

// abort stops the scenario. A rate limit stops all traffic at once, anything
// else still gets its leftovers cleaned up first.
func (s *scenario) abort(ctx context.Context, operation string, err error) {
	if errors.Is(err, common.ErrLimitExceeded) || errors.Is(err, common.ErrRetryable) {
		slog.Error("Rate limited, stopping all traffic. Leftover records:", "created", s.created)
		utils.Fail("rate limited", "operation", operation, "error", err)
	}

	s.cleanup(ctx)
	utils.Fail("custom fields scenario failed", "operation", operation, "error", err)
}

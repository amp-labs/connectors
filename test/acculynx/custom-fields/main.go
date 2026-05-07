package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

// Exercises the read-side custom-fields resolver end-to-end. The connector:
//   1. Fetches definitions from /company-settings/custom-fields
//   2. For each parent record, fetches values from /{contacts|jobs}/{id}/custom-fields
//   3. Joins definitions + values, surfacing fields by their human-readable
//      label slug (e.g. "Customer Preference" → customer_preference) alongside
//      built-in fields, while leaving Raw untouched.
//
// Requires a custom field with a populated value to actually demonstrate the
// flatten — create one via AccuLynx UI (Settings → Custom Fields) and set a
// value on at least one record before running.

// Built-in field names per object — anything else surfaced in Fields is a
// resolved custom field. The connector lowercases field names in the Fields
// map (matching ReadParams.Fields normalization), so these are lowercase too.
//
//nolint:gochecknoglobals
var (
	builtinContactFields = map[string]bool{
		"id": true, "firstname": true, "lastname": true, "salutation": true,
		"crossreference": true, "companyname": true, "mailingaddress": true,
		"billingaddress": true, "phonenumbers": true, "emailaddresses": true,
		"_link": true,
	}

	builtinJobFields = map[string]bool{
		"id": true, "jobname": true, "jobnumber": true, "priority": true,
		"contacts": true, "locationaddress": true, "geolocation": true,
		"tradetypes": true, "jobcategory": true, "worktype": true,
		"leadsource": true, "leaddeadreason": true, "currentmilestone": true,
		"milestonedate": true, "createddate": true, "modifieddate": true,
		"initialappointment": true, "_link": true,
	}
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("=== Test 1: ListObjectMetadata for contacts (includes custom field definitions) ===")
	contactMeta := runMetadata(ctx, conn, "contacts")

	slog.Info("=== Test 2: ListObjectMetadata for jobs (includes custom field definitions) ===")
	jobMeta := runMetadata(ctx, conn, "jobs")

	slog.Info("=== Test 3: Read contacts — verify resolver attaches values per record ===")
	runReadAndFindCustomFields(ctx, conn, "contacts", contactMeta, builtinContactFields)

	slog.Info("=== Test 4: Read jobs — verify resolver attaches values per record ===")
	runReadAndFindCustomFields(ctx, conn, "jobs", jobMeta, builtinJobFields)

	slog.Info("All custom-fields tests completed.")
}

func runMetadata(
	ctx context.Context, conn *acculynx.Connector, objectName string,
) *common.ListObjectMetadataResult {
	res, err := conn.ListObjectMetadata(ctx, []string{objectName})
	if err != nil {
		slog.Error("ListObjectMetadata failed", "object", objectName, "error", err)

		return nil
	}

	objMeta, ok := res.Result[objectName]
	if !ok {
		slog.Warn("Metadata result missing object", "object", objectName)
		utils.DumpJSON(res, os.Stdout)

		return res
	}

	slog.Info("Metadata fetched",
		"object", objectName,
		"displayName", objMeta.DisplayName,
		"totalFields", len(objMeta.Fields))

	utils.DumpJSON(res, os.Stdout)

	return res
}

func runReadAndFindCustomFields(
	ctx context.Context, conn *acculynx.Connector,
	objectName string, meta *common.ListObjectMetadataResult, builtin map[string]bool,
) {
	// Build the Fields set from the metadata: this asks the connector for
	// every field the object exposes, including custom-field slugs surfaced
	// by ListObjectMetadata. Without this, ParseResultFiltered would drop
	// resolved CF values from the output.
	fieldNames := []string{"id"}
	if obj, ok := meta.Result[objectName]; ok {
		for name := range obj.FieldsMap {
			fieldNames = append(fieldNames, name)
		}
	}

	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fieldNames...),
	})
	if err != nil {
		slog.Error("Read failed", "object", objectName, "error", err)

		return
	}

	slog.Info("Read result", "object", objectName, "rows", res.Rows)

	if len(res.Data) == 0 {
		slog.Warn("Empty result — no records to inspect for custom fields", "object", objectName)

		return
	}

	// Scan every record on the page for fields not in the built-in set. Only
	// records that have a CF value set will surface anything — most won't.
	hits := 0

	for _, row := range res.Data {
		extras := make(map[string]any)

		for fieldName, value := range row.Fields {
			// Skip built-ins and skip CF slugs that came back nil (the framework
			// fills any requested field with nil on records that don't have it).
			if builtin[fieldName] || value == nil {
				continue
			}

			extras[fieldName] = value
		}

		if len(extras) == 0 {
			continue
		}

		hits++

		slog.Info("Custom fields resolved and attached to record",
			"object", objectName, "recordId", row.Id, "customFields", extras)
		utils.DumpJSON(row, os.Stdout)
	}

	if hits == 0 {
		slog.Info("No custom field values found on any record in this page",
			"object", objectName, "scanned", len(res.Data))
	}
}

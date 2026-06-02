package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/acculynx"
	testAccuLynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

// AccuLynx supported write operations (every JSON write endpoint AccuLynx
// exposes; file-upload endpoints — documents/measurements/photos-videos —
// require multipart and are intentionally excluded):
//
//   POST  /contacts                                   (Create contact)
//   POST  /jobs                                       (Create job)
//   PUT   /contacts/{id}/custom-fields/{fieldId}      (Update contact CF value)
//   PUT   /jobs/{id}/custom-fields/{fieldId}          (Update job CF value)
//   PUT   /jobs/{id}/initial-appointment              (Set appointment slot)
//   PUT   /jobs/{id}/insurance/insurance-company      (Set insurance carrier)
//   POST  /jobs/{id}/representatives/ar-owner         (Assign AR owner)
//   POST  /jobs/{id}/representatives/sales-owner      (Assign sales owner)
//   POST  /jobs/{id}/representatives/company          (Assign company rep)
//   POST  /contacts/{id}/phone-numbers                (Add phone number)
//   POST  /contacts/{id}/logs                         (Add log entry)
//   POST  /jobs/external-references                   (Cross-system mapping)
//   POST  /jobs/{id}/messages                         (Send message)
//   POST  /jobs/{id}/messages/{messageId}/replies     (Reply to message)
//   POST  /jobs/{id}/payments/expense                 (Add expense)
//   POST  /jobs/{id}/payments/paid                    (Mark paid)
//   POST  /jobs/{id}/payments/received                (Mark received)

// Constants pulled from the AccuLynx sandbox; replace if testing against a
// different account.
const (
	existingJobID           = "9ecc68c2-9beb-4b8f-a4b5-6f4e52a41d75"
	customerContactTypeID   = "52ba94c5-3ecf-4e7f-90cd-a91de12a72f5" // "Customer" type
	existingContactIDForJob = "f5d6d7fe-fd00-ef11-9149-3cecef1c5971" // contact linked to existingJobID
	existingUserID          = "0c9b5100-bf4d-41d3-87c8-3729269baeef" // integrations user
)

func main() { //nolint:funlen,cyclop
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := testAccuLynx.GetAccuLynxConnector(ctx)

	slog.Info("=== Test 1: Create Contact (POST /contacts) ===")
	contactID, err := testCreateContact(ctx, conn)
	logResult("Create contact", contactID, err)

	slog.Info("=== Test 2: Create Job (POST /jobs) ===")
	contactIDForJob := contactID
	if contactIDForJob == "" {
		// Falls back to a known sandbox contact when Test 1 didn't return one.
		contactIDForJob = existingContactIDForJob
	}
	jobID, err := testCreateJob(ctx, conn, contactIDForJob)
	logResult("Create job", jobID, err)

	// Placeholder fieldId — sandbox has no custom fields, so these will 404.
	slog.Info("=== Test 3: PUT Contact Custom Field ===")
	if contactID != "" {
		runWrite(ctx, conn, "contacts/custom-fields", "00000000-0000-0000-0000-000000000001", map[string]any{
			"contactId": contactID,
			"fieldType": "Text",
			"values":    []string{"ampersand-write-test"},
		})
	}

	slog.Info("=== Test 4: PUT Job Custom Field ===")
	runWrite(ctx, conn, "jobs/custom-fields", "00000000-0000-0000-0000-000000000001", map[string]any{
		"jobId":     existingJobID,
		"fieldType": "Text",
		"values":    []string{"ampersand-write-test"},
	})

	slog.Info("=== Test 5: PUT Initial Appointment ===")
	runWrite(ctx, conn, "jobs/initial-appointment", "", map[string]any{
		"jobId":     existingJobID,
		"startDate": "2026-06-22T18:47:10Z",
		"endDate":   "2026-06-22T19:47:10Z",
		"notes":     "Updated via Ampersand live write test",
	})

	slog.Info("=== Test 6: PUT Insurance Company (passes name instead of ID) ===")
	runWrite(ctx, conn, "jobs/insurance/insurance-company", "", map[string]any{
		"jobId":                existingJobID,
		"insuranceCompanyName": "Ampersand Test Insurance",
	})

	slog.Info("=== Test 7: POST AR Owner ===")
	runWrite(ctx, conn, "jobs/representatives/ar-owner", "", map[string]any{
		"jobId": existingJobID,
		"id":    existingUserID,
	})

	slog.Info("=== Test 8: POST Sales Owner ===")
	runWrite(ctx, conn, "jobs/representatives/sales-owner", "", map[string]any{
		"jobId": existingJobID,
		"id":    existingUserID,
	})

	slog.Info("=== Test 9: POST Company Representative ===")
	runWrite(ctx, conn, "jobs/representatives/company", "", map[string]any{
		"jobId": existingJobID,
		"id":    existingUserID,
	})

	if contactID != "" {
		slog.Info("=== Test 10: POST Contact Phone Number ===")
		runWrite(ctx, conn, "contacts/phone-numbers", "", map[string]any{
			"contactId": contactID,
			"number":    fmt.Sprintf("555%07d", gofakeit.Number(1000000, 9999999)),
			"type":      "Mobile",
			"primary":   false,
		})

		slog.Info("=== Test 11: POST Contact Log ===")
		runWrite(ctx, conn, "contacts/logs", "", map[string]any{
			"contactId": contactID,
			"logDate":   time.Now().UTC().Format(time.RFC3339),
			"type":      "PhoneCall",
			"note":      "Ampersand live write test log entry",
		})
	}

	slog.Info("=== Test 12: POST Job External Reference ===")
	runWrite(ctx, conn, "jobs/external-references", "", map[string]any{
		"jobId":     existingJobID,
		"source":    "ampersand-test",
		"projectId": fmt.Sprintf("amp-test-%d", time.Now().Unix()),
	})

	slog.Info("=== Test 13: POST Job Message ===")
	runWrite(ctx, conn, "jobs/messages", "", map[string]any{
		"jobId":   existingJobID,
		"message": fmt.Sprintf("Ampersand live write test message %s", time.Now().Format(time.RFC3339)),
	})

	slog.Info("=== Test 14: POST Message Reply (will 404 unless messageId exists) ===")
	runWrite(ctx, conn, "jobs/messages/replies", "", map[string]any{
		"jobId":     existingJobID,
		"messageId": "00000000-0000-0000-0000-000000000001",
		"message":   "Ampersand test reply",
	})

	slog.Info("=== Test 15: POST Payment Expense ===")
	runWrite(ctx, conn, "jobs/payments/expense", "", map[string]any{
		"jobId":  existingJobID,
		"to":     "Ampersand Test Vendor",
		"amount": 100,
		"notes":  "Live write test - expense",
	})

	slog.Info("=== Test 16: POST Payment Paid ===")
	runWrite(ctx, conn, "jobs/payments/paid", "", map[string]any{
		"jobId":       existingJobID,
		"to":          "Ampersand Test Vendor",
		"amount":      200,
		"paymentDate": "2026-05-19T00:00:00Z",
		"isPaid":      true,
	})

	slog.Info("=== Test 17: POST Payment Received ===")
	runWrite(ctx, conn, "jobs/payments/received", "", map[string]any{
		"jobId":       existingJobID,
		"from":        "Ampersand Test Client",
		"amount":      300,
		"paymentDate": "2026-05-19T00:00:00Z",
	})

	slog.Info("All write tests completed.")
}

// runWrite invokes Write with the given params and logs the outcome. Errors are
// recorded as warnings (not fatal) so subsequent tests still run — useful when
// a sandbox is missing referenced IDs (insurance companies, account types,
// pre-existing messages, etc.) but the connector code path itself is what we
// want to verify.
func runWrite(
	ctx context.Context, conn *acculynx.Connector,
	objectName, recordID string, recordData map[string]any,
) {
	params := common.WriteParams{
		ObjectName: objectName,
		RecordId:   recordID,
		RecordData: recordData,
	}

	slog.Info("Write", "object", objectName, "recordId", recordID, "data", recordData)

	res, err := conn.Write(ctx, params)
	if err != nil {
		slog.Warn("Write returned an error (may be expected if sandbox is missing referenced IDs)",
			"object", objectName, "error", err)

		return
	}

	slog.Info("✅ Write succeeded", "object", objectName, "recordId", res.RecordId, "success", res.Success)
	utils.DumpJSON(res, os.Stdout)
}

func logResult(action, recordID string, err error) {
	if err != nil {
		slog.Error(action+" failed", "error", err)

		return
	}

	slog.Info("✅ "+action+" succeeded", "recordId", recordID)
}

func testCreateContact(ctx context.Context, conn *acculynx.Connector) (string, error) {
	recordData := map[string]any{
		"contactTypeIds": []string{customerContactTypeID},
		"firstName":      gofakeit.FirstName(),
		"lastName":       gofakeit.LastName(),
		"mailingAddress": map[string]any{
			"street1": gofakeit.StreetName(),
			"city":    gofakeit.City(),
			"zipCode": gofakeit.Zip(),
			"state":   map[string]any{"id": 46}, // Virginia (sandbox-known)
			"country": map[string]any{"id": 1},  // United States
		},
	}

	slog.Info("Creating contact", "data", recordData)

	params := common.WriteParams{
		ObjectName: "contacts",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("create contact: %w", err)
	}

	slog.Info("Contact create result", "recordId", res.RecordId, "success", res.Success)
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

func testCreateJob(ctx context.Context, conn *acculynx.Connector, contactID string) (string, error) {
	recordData := map[string]any{
		"contact": map[string]any{"id": contactID},
		"notes":   fmt.Sprintf("Ampersand Test Job %s", time.Now().Format("2006-01-02T15:04:05Z")),
		"locationAddress": map[string]any{
			"street1": gofakeit.StreetName(),
			"city":    gofakeit.City(),
			"zipCode": gofakeit.Zip(),
			"state":   "VA",
			"country": "US",
		},
	}

	slog.Info("Creating job", "data", recordData)

	params := common.WriteParams{
		ObjectName: "jobs",
		RecordData: recordData,
	}

	res, err := conn.Write(ctx, params)
	if err != nil {
		return "", fmt.Errorf("create job: %w", err)
	}

	slog.Info("Job create result", "recordId", res.RecordId, "success", res.Success)
	utils.DumpJSON(res, os.Stdout)

	return res.RecordId, nil
}

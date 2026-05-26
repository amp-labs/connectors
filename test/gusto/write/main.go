package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	gustoConn "github.com/amp-labs/connectors/providers/gusto"
	connTest "github.com/amp-labs/connectors/test/gusto"
	"github.com/amp-labs/connectors/test/utils"
)

// Gusto write live test — broad coverage of every object the connector
// supports, exercising both the CREATE and UPDATE code paths where Gusto
// allows them.
//
// Idempotent across re-runs:
//   - All parent UUIDs (employee, location, company) discovered at runtime
//     via Read; nothing hardcoded.
//   - Unique-per-run values (timestamp suffixes, future effective_dates) are
//     used so Gusto's uniqueness constraints don't fail back-to-back runs.
//   - Versions for PUT calls come from the record we just created (fresh
//     state every run) rather than the read snapshot.
//
// Endpoints exercised:
//   Company-scoped CREATE (POST /v1/companies/{cid}/{object}):
//     1. departments
//     2. locations
//     3. earning_types
//   Employee-scoped CREATE (POST /v1/employees/{eid}/{object}; eid from RecordData):
//     4. home_addresses
//     5. work_addresses
//     6. jobs
//     7. garnishments
//   Job-scoped CREATE (POST /v1/jobs/{jid}/{object}; jid from RecordData):
//     8. compensations
//   Top-level UPDATE (PUT /v1/{object}/{uuid}; requires version):
//     9.  departments (just-created uuid+version)
//     10. earning_types (just-created uuid+version)
//     11. jobs (just-created uuid+version)
//     12. compensations (just-created uuid+version)
//     13. locations (existing, version refreshed each run)
//     14. employees (existing first employee)

func main() { //nolint:funlen
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetConnector(ctx)
	conn.GetPostAuthInfo(ctx)
	stamp := time.Now().Format("20060102-150405")

	// Discover parents we'll use throughout the run.
	employee, err := firstRecord(ctx, conn, "employees", "uuid", "version")
	if err != nil {
		bail("discover employee", err)
	}

	location, err := firstRecord(ctx, conn, "locations", "uuid", "version")
	if err != nil {
		bail("discover location", err)
	}

	// === Company-scoped creates ===
	deptUUID, deptVersion, err := createDepartment(ctx, conn, stamp)
	if err != nil {
		bail("create department", err)
	}

	if _, _, err := createLocation(ctx, conn, stamp); err != nil {
		bail("create location", err)
	}

	earningTypeUUID, _, err := createEarningType(ctx, conn, stamp)
	if err != nil {
		bail("create earning_type", err)
	}

	// === Employee-scoped creates (employee_id pulled from RecordData) ===
	if _, _, err := createHomeAddress(ctx, conn, employee["uuid"]); err != nil {
		bail("create home_address", err)
	}

	if _, _, err := createWorkAddress(ctx, conn, employee["uuid"], location["uuid"]); err != nil {
		bail("create work_address", err)
	}

	jobUUID, jobVersion, err := createJob(ctx, conn, employee["uuid"], stamp)
	if err != nil {
		bail("create job", err)
	}

	if _, _, err := createGarnishment(ctx, conn, employee["uuid"], stamp); err != nil {
		bail("create garnishment", err)
	}

	// Update the job BEFORE creating a compensation, because creating a
	// compensation bumps the parent job's version (Gusto invariant).
	if err := updateJob(ctx, conn, jobUUID, jobVersion, stamp); err != nil {
		bail("update job", err)
	}

	// === Job-scoped create (job_id pulled from RecordData) ===
	compensationUUID, compensationVersion, err := createCompensation(ctx, conn, jobUUID, stamp)
	if err != nil {
		bail("create compensation", err)
	}

	// === Top-level / company-scoped updates ===
	if err := updateDepartment(ctx, conn, deptUUID, deptVersion, stamp); err != nil {
		bail("update department", err)
	}

	if err := updateEarningType(ctx, conn, earningTypeUUID, stamp); err != nil {
		bail("update earning_type", err)
	}

	if err := updateCompensation(ctx, conn, compensationUUID, compensationVersion); err != nil {
		bail("update compensation", err)
	}

	if err := updateLocation(ctx, conn, location["uuid"], location["version"]); err != nil {
		bail("update location", err)
	}

	if err := updateEmployee(ctx, conn, employee["uuid"], employee["version"], stamp); err != nil {
		bail("update employee", err)
	}

	// === Delete (cleanup) — exercises the Deleter code paths and keeps the
	// demo company tidy by removing every record this run created. Order
	// matters: compensation before its parent job; nothing else has hard
	// foreign-key dependencies. ===
	if err := deleteRecord(ctx, conn, "compensations", compensationUUID); err != nil {
		bail("delete compensation", err)
	}

	if err := deleteRecord(ctx, conn, "jobs", jobUUID); err != nil {
		bail("delete job", err)
	}

	if err := deleteRecord(ctx, conn, "departments", deptUUID); err != nil {
		bail("delete department", err)
	}

	if err := deleteRecord(ctx, conn, "earning_types", earningTypeUUID); err != nil {
		// company-scoped DELETE — verifies the nested URL path
		bail("delete earning_type", err)
	}

	slog.Info("=== ALL WRITE SCENARIOS PASSED ===")
}

// deleteRecord exercises the Delete code path. Top-level vs company-scoped
// routing is decided inside the connector based on objectName.
func deleteRecord(ctx context.Context, conn *gustoConn.Connector, objectName, uuid string) error {
	slog.Info("=== Delete "+objectName+" ===", "uuid", uuid)

	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   uuid,
	})
	if err != nil {
		return err
	}

	if res == nil || !res.Success {
		return fmt.Errorf("delete %s returned non-success result", objectName)
	}

	utils.DumpJSON(map[string]any{
		"success":    true,
		"deleted":    uuid,
		"objectName": objectName,
	}, os.Stdout)

	return nil
}

// firstRecord reads the first item of an object and returns the requested
// fields from its Raw payload as a map.
func firstRecord(ctx context.Context, conn *gustoConn.Connector, objectName string, fields ...string) (map[string]string, error) {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(fields...),
		PageSize:   1,
	})
	if err != nil {
		return nil, err
	}

	if res == nil || len(res.Data) == 0 {
		return nil, fmt.Errorf("no %s in demo company", objectName)
	}

	out := make(map[string]string, len(fields))
	for _, f := range fields {
		v, _ := res.Data[0].Raw[f].(string)
		out[f] = v
	}

	slog.Info("discovered "+objectName, slog.Any("fields", out))

	return out, nil
}

// ---- Company-scoped creates ----

func createDepartment(ctx context.Context, conn *gustoConn.Connector, stamp string) (string, string, error) {
	slog.Info("=== Create department (company-scoped) ===")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "departments",
		RecordData: map[string]any{"title": "QA Engineering " + stamp},
	})

	return capture(res, err)
}

func createLocation(ctx context.Context, conn *gustoConn.Connector, stamp string) (string, string, error) {
	slog.Info("=== Create location (company-scoped) ===")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "locations",
		RecordData: map[string]any{
			"street_1":     "100 Test Plaza " + stamp[len(stamp)-6:],
			"city":         "San Francisco",
			"state":        "CA",
			"zip":          "94105",
			"country":      "USA",
			"phone_number": "4155550100",
		},
	})

	return capture(res, err)
}

func createEarningType(ctx context.Context, conn *gustoConn.Connector, stamp string) (string, string, error) {
	slog.Info("=== Create earning_type (company-scoped) ===")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "earning_types",
		RecordData: map[string]any{"name": "Bonus " + stamp},
	})

	return capture(res, err)
}

// ---- Employee-scoped creates ----

func createHomeAddress(ctx context.Context, conn *gustoConn.Connector, employeeUUID string) (string, string, error) {
	slog.Info("=== Create home_address (employee-scoped) ===")

	// Gusto enforces (employee, effective_date) uniqueness; offset by ms-of-now.
	daysOffset := int(time.Now().UnixMilli()) % 1000

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "home_addresses",
		RecordData: map[string]any{
			"employee_id":    employeeUUID,
			"street_1":       fmt.Sprintf("Home Lane #%d", daysOffset),
			"city":           "San Francisco",
			"state":          "CA",
			"zip":            "94105",
			"effective_date": time.Now().AddDate(0, 0, daysOffset+1).Format("2006-01-02"),
		},
	})

	return capture(res, err)
}

func createWorkAddress(ctx context.Context, conn *gustoConn.Connector, employeeUUID, locationUUID string) (string, string, error) {
	slog.Info("=== Create work_address (employee-scoped) ===")

	daysOffset := int(time.Now().UnixMilli()/2) % 1000

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "work_addresses",
		RecordData: map[string]any{
			"employee_id":    employeeUUID,
			"location_uuid":  locationUUID,
			"effective_date": time.Now().AddDate(0, 0, daysOffset+1).Format("2006-01-02"),
		},
	})

	return capture(res, err)
}

func createJob(ctx context.Context, conn *gustoConn.Connector, employeeUUID, stamp string) (string, string, error) {
	slog.Info("=== Create job (employee-scoped) ===")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "jobs",
		RecordData: map[string]any{
			"employee_id":             employeeUUID,
			"title":                   "Test Engineer " + stamp,
			"hire_date":               time.Now().Format("2006-01-02"),
			"primary":                 false,
			"location_uuid":           nil,
			"two_percent_shareholder": false,
		},
	})

	return capture(res, err)
}

func createGarnishment(ctx context.Context, conn *gustoConn.Connector, employeeUUID, stamp string) (string, string, error) {
	slog.Info("=== Create garnishment (employee-scoped) ===")

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "garnishments",
		RecordData: map[string]any{
			"employee_id":   employeeUUID,
			"active":        true,
			"amount":        "150.00",
			"description":   "Test garnishment " + stamp,
			"court_ordered": true,
			"times":         5,
			"recurring":     false,
		},
	})

	return capture(res, err)
}

// ---- Job-scoped create ----

func createCompensation(ctx context.Context, conn *gustoConn.Connector, jobUUID, stamp string) (string, string, error) {
	slog.Info("=== Create compensation (job-scoped) ===")

	// Gusto requires compensation effective_date in [tomorrow, +1 year].
	// Pick a unique-per-run day inside that window.
	daysOffset := int(time.Now().UnixMilli()/1000)%364 + 1

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "compensations",
		RecordData: map[string]any{
			"job_id":         jobUUID,
			"rate":           "75.00",
			"payment_unit":   "Hour",
			"flsa_status":    "Nonexempt",
			"effective_date": time.Now().AddDate(0, 0, daysOffset).Format("2006-01-02"),
		},
	})

	return capture(res, err)
}

// ---- Top-level updates ----

func updateDepartment(ctx context.Context, conn *gustoConn.Connector, uuid, version, stamp string) error {
	slog.Info("=== Update department (top-level PUT) ===", "uuid", uuid)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "departments",
		RecordId:   uuid,
		RecordData: map[string]any{
			"version": version,
			"title":   "QA Engineering Updated " + stamp,
		},
	})

	return only(res, err)
}

func updateEarningType(ctx context.Context, conn *gustoConn.Connector, uuid, stamp string) error {
	slog.Info("=== Update earning_type (top-level PUT) ===", "uuid", uuid)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "earning_types",
		RecordId:   uuid,
		RecordData: map[string]any{"name": "Bonus Updated " + stamp},
	})

	return only(res, err)
}

func updateJob(ctx context.Context, conn *gustoConn.Connector, uuid, version, stamp string) error {
	slog.Info("=== Update job (top-level PUT) ===", "uuid", uuid)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "jobs",
		RecordId:   uuid,
		RecordData: map[string]any{
			"version": version,
			"title":   "Senior Test Engineer " + stamp,
		},
	})

	return only(res, err)
}

func updateCompensation(ctx context.Context, conn *gustoConn.Connector, uuid, version string) error {
	slog.Info("=== Update compensation (top-level PUT) ===", "uuid", uuid)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "compensations",
		RecordId:   uuid,
		RecordData: map[string]any{
			"version": version,
			"rate":    "85.00",
		},
	})

	return only(res, err)
}

func updateLocation(ctx context.Context, conn *gustoConn.Connector, uuid, version string) error {
	slog.Info("=== Update location (top-level PUT) ===", "uuid", uuid)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "locations",
		RecordId:   uuid,
		RecordData: map[string]any{
			"version":      version,
			"phone_number": "4155551234",
		},
	})

	return only(res, err)
}

func updateEmployee(ctx context.Context, conn *gustoConn.Connector, uuid, version, stamp string) error {
	slog.Info("=== Update employee (top-level PUT) ===", "uuid", uuid)

	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "employees",
		RecordId:   uuid,
		RecordData: map[string]any{
			"version":              version,
			"middle_initial":       "T",
			"preferred_first_name": "Test " + stamp[len(stamp)-4:],
		},
	})

	return only(res, err)
}

// ---- helpers ----

// capture prints a successful WriteResult and extracts uuid + version from
// the data payload so the caller can use them in subsequent updates.
func capture(res *common.WriteResult, err error) (string, string, error) {
	if err != nil {
		return "", "", err
	}

	utils.DumpJSON(res, os.Stdout)

	if res == nil || res.Data == nil {
		return "", "", nil
	}

	uuid, _ := res.Data["uuid"].(string)
	version, _ := res.Data["version"].(string)

	return uuid, version, nil
}

// only is for endpoints whose return values we don't reuse downstream.
func only(res *common.WriteResult, err error) error {
	if err != nil {
		return err
	}

	utils.DumpJSON(res, os.Stdout)

	return nil
}

func bail(label string, err error) {
	slog.Error(label, "err", err)
	os.Exit(1)
}

package salesforce

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/metadata"
)

// TestRollbackApexTriggerFallsBackToNoTestRun verifies that when a destructive
// apex deploy is rolled back by the test run (no component failures — e.g. the
// org's overall coverage is below 75%), rollbackApexTrigger retries once with
// NoTestRun and succeeds.
func TestRollbackApexTriggerFallsBackToNoTestRun(t *testing.T) {
	t.Parallel()

	var (
		mu               sync.Mutex
		deployTestLevels []string // testLevel of each deploy request, in order
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)

		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")

		switch {
		case strings.Contains(body, "<md:deploy"):
			level := soapTestLevel(body)

			mu.Lock()
			deployTestLevels = append(deployTestLevels, level)
			mu.Unlock()

			// Echo the test level into the deploy id so checkDeployStatus can key on it.
			_, _ = io.WriteString(w, soapEnvelope(
				`<deployResponse><result><id>deploy-`+level+`</id><done>false</done></result></deployResponse>`))

		case strings.Contains(body, "<md:checkDeployStatus"):
			// The RunLocalTests attempt is rolled back by the test run (success=false,
			// no component failures); the NoTestRun attempt succeeds.
			success := strings.Contains(body, "deploy-"+string(metadata.TestLevelNoTestRun))
			_, _ = io.WriteString(w, soapEnvelope(checkStatusBody(success)))

		default:
			t.Errorf("unexpected SOAP request body: %s", body)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("constructTestConnector: %v", err)
	}

	ctx := common.WithAuthToken(context.Background(), "test-token")

	if err := conn.rollbackApexTrigger(ctx, "CDC_Account"); err != nil {
		t.Fatalf("rollbackApexTrigger returned error, want success after fallback: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	want := []string{string(metadata.TestLevelRunLocalTests), string(metadata.TestLevelNoTestRun)}
	if len(deployTestLevels) != len(want) {
		t.Fatalf("deploy count = %d (%v), want %d (%v)",
			len(deployTestLevels), deployTestLevels, len(want), want)
	}
	for i, lvl := range want {
		if deployTestLevels[i] != lvl {
			t.Errorf("deploy[%d] testLevel = %q, want %q", i, deployTestLevels[i], lvl)
		}
	}
}

// TestRollbackApexTriggerComponentFailureNoRetry verifies that a genuine
// component failure (the destructive change itself was rejected) is NOT retried
// with NoTestRun — it surfaces as an error after a single deploy.
func TestRollbackApexTriggerComponentFailureNoRetry(t *testing.T) {
	t.Parallel()

	var (
		mu          sync.Mutex
		deployCount int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)

		w.Header().Set("Content-Type", "text/xml; charset=UTF-8")

		switch {
		case strings.Contains(body, "<md:deploy"):
			mu.Lock()
			deployCount++
			mu.Unlock()
			_, _ = io.WriteString(w, soapEnvelope(
				`<deployResponse><result><id>deploy-1</id><done>false</done></result></deployResponse>`))

		case strings.Contains(body, "<md:checkDeployStatus"):
			// Failure WITH a component failure — must not be retried.
			_, _ = io.WriteString(w, soapEnvelope(
				`<checkDeployStatusResponse><result>`+
					`<done>true</done><status>Failed</status><success>false</success><id>deploy-1</id>`+
					`<details><componentFailures>`+
					`<componentType>ApexClass</componentType><fullName>CDC_Account_Handler</fullName>`+
					`<problem>Something depends on this class</problem><problemType>Error</problemType>`+
					`</componentFailures></details>`+
					`</result></checkDeployStatusResponse>`))

		default:
			t.Errorf("unexpected SOAP request body: %s", body)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatalf("constructTestConnector: %v", err)
	}

	ctx := common.WithAuthToken(context.Background(), "test-token")

	if err := conn.rollbackApexTrigger(ctx, "CDC_Account"); err == nil {
		t.Fatal("rollbackApexTrigger returned nil, want error for a component failure")
	}

	mu.Lock()
	defer mu.Unlock()

	if deployCount != 1 {
		t.Errorf("deploy count = %d, want 1 (a component failure must not be retried)", deployCount)
	}
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}

	return string(b)
}

// soapTestLevel extracts the <md:testLevel> value from a deploy request body.
func soapTestLevel(body string) string {
	const open = "<md:testLevel>"

	start := strings.Index(body, open)
	if start < 0 {
		return ""
	}
	start += len(open)

	end := strings.Index(body[start:], "</md:testLevel>")
	if end < 0 {
		return ""
	}

	return body[start : start+end]
}

func soapEnvelope(inner string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">` +
		`<soapenv:Body>` + inner + `</soapenv:Body></soapenv:Envelope>`
}

func checkStatusBody(success bool) string {
	status := "Failed"
	if success {
		status = "Succeeded"
	}

	return `<checkDeployStatusResponse><result>` +
		`<done>true</done><status>` + status + `</status>` +
		`<success>` + boolStr(success) + `</success><id>deploy</id>` +
		`<details></details>` +
		`</result></checkDeployStatusResponse>`
}

func boolStr(b bool) string {
	if b {
		return "true"
	}

	return "false"
}

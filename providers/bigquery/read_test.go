package bigquery

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/amp-labs/connectors/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// Token serialization
// =============================================================================

func TestTokenRoundTrip(t *testing.T) {
	t.Parallel()

	original := &readSessionToken{
		SessionName:  "projects/p/locations/us/sessions/abc123",
		ActiveStreams: 2,
		Streams: []streamState{
			{Name: "stream-0", Offset: 1000, Done: false},
			{Name: "stream-1", Offset: 500, Done: false},
			{Name: "stream-2", Offset: 0, Done: true},
		},
		IsBackfill:  true,
		WindowStart: "2024-01-01T00:00:00Z",
		WindowEnd:   "2024-01-31T00:00:00Z",
	}

	encoded, err := encodeReadSessionToken(original)
	if err != nil {
		t.Fatalf("encodeReadSessionToken() error = %v", err)
	}

	if encoded == "" {
		t.Fatal("encodeReadSessionToken() returned empty string")
	}

	// Verify it's valid base64.
	decoded, err := parseReadSessionToken(encoded)
	if err != nil {
		t.Fatalf("parseReadSessionToken() error = %v", err)
	}

	if decoded.SessionName != original.SessionName {
		t.Errorf("SessionName = %q, want %q", decoded.SessionName, original.SessionName)
	}

	if decoded.ActiveStreams != original.ActiveStreams {
		t.Errorf("ActiveStreams = %d, want %d", decoded.ActiveStreams, original.ActiveStreams)
	}

	if len(decoded.Streams) != len(original.Streams) {
		t.Fatalf("len(Streams) = %d, want %d", len(decoded.Streams), len(original.Streams))
	}

	for i, s := range decoded.Streams {
		if s.Name != original.Streams[i].Name {
			t.Errorf("Streams[%d].Name = %q, want %q", i, s.Name, original.Streams[i].Name)
		}

		if s.Offset != original.Streams[i].Offset {
			t.Errorf("Streams[%d].Offset = %d, want %d", i, s.Offset, original.Streams[i].Offset)
		}

		if s.Done != original.Streams[i].Done {
			t.Errorf("Streams[%d].Done = %v, want %v", i, s.Done, original.Streams[i].Done)
		}
	}

	if decoded.IsBackfill != original.IsBackfill {
		t.Errorf("IsBackfill = %v, want %v", decoded.IsBackfill, original.IsBackfill)
	}

	if decoded.WindowStart != original.WindowStart {
		t.Errorf("WindowStart = %q, want %q", decoded.WindowStart, original.WindowStart)
	}

	if decoded.WindowEnd != original.WindowEnd {
		t.Errorf("WindowEnd = %q, want %q", decoded.WindowEnd, original.WindowEnd)
	}
}

func TestTokenOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	// Incremental read token — no window fields.
	token := &readSessionToken{
		SessionName:  "projects/p/locations/us/sessions/xyz",
		ActiveStreams: 1,
		Streams: []streamState{
			{Name: "stream-0", Offset: 0, Done: false},
		},
	}

	encoded, err := encodeReadSessionToken(token)
	if err != nil {
		t.Fatalf("encodeReadSessionToken() error = %v", err)
	}

	// Decode the base64 to inspect the JSON.
	jsonBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64.DecodeString() error = %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Window fields should be omitted for incremental reads.
	for _, key := range []string{"windowStart", "windowEnd", "isBackfill"} {
		if _, exists := raw[key]; exists {
			t.Errorf("expected %q to be omitted from JSON, but it was present", key)
		}
	}
}

func TestParseInvalidToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "not base64", input: "not-valid-base64!!!"},
		{name: "valid base64 but not JSON", input: base64.StdEncoding.EncodeToString([]byte("not json"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseReadSessionToken(tt.input)
			if err == nil {
				t.Error("parseReadSessionToken() expected error, got nil")
			}
		})
	}
}

// =============================================================================
// Token resolution (resolveToken)
// =============================================================================

func TestResolveToken_NextPage(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	original := &readSessionToken{
		SessionName:  "session-123",
		ActiveStreams: 1,
		Streams:      []streamState{{Name: "s0", Offset: 50000}},
		IsBackfill:   true,
		WindowStart:  "2024-01-01T00:00:00Z",
		WindowEnd:    "2024-01-31T00:00:00Z",
	}

	encoded, _ := encodeReadSessionToken(original)

	token, err := c.resolveToken(common.ReadParams{
		NextPage: common.NextPageToken(encoded),
	})
	if err != nil {
		t.Fatalf("resolveToken() error = %v", err)
	}

	if token.SessionName != "session-123" {
		t.Errorf("SessionName = %q, want %q", token.SessionName, "session-123")
	}

	if !token.IsBackfill {
		t.Error("expected IsBackfill = true")
	}
}

func TestResolveToken_Incremental(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	since := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	token, err := c.resolveToken(common.ReadParams{
		Since: since,
	})
	if err != nil {
		t.Fatalf("resolveToken() error = %v", err)
	}

	if token.IsBackfill {
		t.Error("expected IsBackfill = false for incremental read")
	}

	if token.SessionName != "" {
		t.Errorf("expected empty SessionName, got %q", token.SessionName)
	}

	if token.WindowStart != "" {
		t.Errorf("expected empty WindowStart, got %q", token.WindowStart)
	}
}

func TestResolveToken_Backfill(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	token, err := c.resolveToken(common.ReadParams{})
	if err != nil {
		t.Fatalf("resolveToken() error = %v", err)
	}

	if !token.IsBackfill {
		t.Error("expected IsBackfill = true for backfill")
	}

	if token.WindowStart == "" {
		t.Error("expected WindowStart to be set")
	}

	if token.WindowEnd == "" {
		t.Error("expected WindowEnd to be set")
	}

	// WindowStart should be the epoch.
	start, err := time.Parse(time.RFC3339, token.WindowStart)
	if err != nil {
		t.Fatalf("failed to parse WindowStart: %v", err)
	}

	if !start.Equal(backfillEpoch) {
		t.Errorf("WindowStart = %v, want %v", start, backfillEpoch)
	}

	// WindowEnd should be epoch + 30 days.
	end, err := time.Parse(time.RFC3339, token.WindowEnd)
	if err != nil {
		t.Fatalf("failed to parse WindowEnd: %v", err)
	}

	expectedEnd := backfillEpoch.Add(backfillWindowSize)
	if !end.Equal(expectedEnd) {
		t.Errorf("WindowEnd = %v, want %v", end, expectedEnd)
	}
}

// =============================================================================
// Window advancement
// =============================================================================

func TestAdvanceWindow(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	token := &readSessionToken{
		IsBackfill:  true,
		WindowStart: "2024-01-01T00:00:00Z",
		WindowEnd:   "2024-01-31T00:00:00Z",
	}

	advanced, err := c.advanceWindow(token)
	if err != nil {
		t.Fatalf("advanceWindow() error = %v", err)
	}

	if advanced.WindowStart != "2024-01-31T00:00:00Z" {
		t.Errorf("WindowStart = %q, want %q", advanced.WindowStart, "2024-01-31T00:00:00Z")
	}

	// WindowEnd should be 30 days after new start.
	expectedEnd := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC).Add(backfillWindowSize)

	end, err := time.Parse(time.RFC3339, advanced.WindowEnd)
	if err != nil {
		t.Fatalf("failed to parse WindowEnd: %v", err)
	}

	if !end.Equal(expectedEnd) {
		t.Errorf("WindowEnd = %v, want %v", end, expectedEnd)
	}

	if !advanced.IsBackfill {
		t.Error("expected IsBackfill to remain true")
	}

	// Session should be cleared for new window.
	if advanced.SessionName != "" {
		t.Errorf("expected empty SessionName, got %q", advanced.SessionName)
	}

	if advanced.Streams != nil {
		t.Error("expected nil Streams")
	}
}

func TestAdvanceWindow_PastNow(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	// Set window end to the future — should signal backfill complete.
	futureEnd := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	token := &readSessionToken{
		IsBackfill:  true,
		WindowStart: "2025-12-01T00:00:00Z",
		WindowEnd:   futureEnd,
	}

	_, err := c.advanceWindow(token)
	if err == nil {
		t.Error("advanceWindow() expected error for future window end, got nil")
	}
}

// =============================================================================
// Row restriction building
// =============================================================================

func TestBuildRowRestriction_Backfill(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	token := &readSessionToken{
		IsBackfill:  true,
		WindowStart: "2024-01-01T00:00:00Z",
		WindowEnd:   "2024-01-31T00:00:00Z",
	}

	restriction := c.buildRowRestriction(token, common.ReadParams{})

	expected := "updated_at >= TIMESTAMP('2024-01-01T00:00:00Z') AND updated_at < TIMESTAMP('2024-01-31T00:00:00Z')"
	if restriction != expected {
		t.Errorf("buildRowRestriction() = %q, want %q", restriction, expected)
	}
}

func TestBuildRowRestriction_Incremental(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "modified_at"}

	since := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	until := time.Date(2024, 6, 15, 10, 35, 0, 0, time.UTC)

	token := &readSessionToken{IsBackfill: false}

	restriction := c.buildRowRestriction(token, common.ReadParams{
		Since: since,
		Until: until,
	})

	expected := fmt.Sprintf(
		"modified_at >= TIMESTAMP('%s') AND modified_at < TIMESTAMP('%s')",
		since.UTC().Format(time.RFC3339),
		until.UTC().Format(time.RFC3339),
	)
	if restriction != expected {
		t.Errorf("buildRowRestriction() = %q, want %q", restriction, expected)
	}
}

func TestBuildRowRestriction_WithFilter(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	token := &readSessionToken{
		IsBackfill:  true,
		WindowStart: "2024-01-01T00:00:00Z",
		WindowEnd:   "2024-01-31T00:00:00Z",
	}

	restriction := c.buildRowRestriction(token, common.ReadParams{
		Filter: "country_code = 'US'",
	})

	expected := "updated_at >= TIMESTAMP('2024-01-01T00:00:00Z') AND updated_at < TIMESTAMP('2024-01-31T00:00:00Z') AND country_code = 'US'"
	if restriction != expected {
		t.Errorf("buildRowRestriction() = %q, want %q", restriction, expected)
	}
}

func TestBuildRowRestriction_IncrementalSinceOnly(t *testing.T) {
	t.Parallel()

	c := &Connector{timestampColumn: "updated_at"}

	since := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	token := &readSessionToken{IsBackfill: false}

	restriction := c.buildRowRestriction(token, common.ReadParams{
		Since: since,
	})

	expected := fmt.Sprintf("updated_at >= TIMESTAMP('%s')", since.UTC().Format(time.RFC3339))
	if restriction != expected {
		t.Errorf("buildRowRestriction() = %q, want %q", restriction, expected)
	}
}

// =============================================================================
// Session expiry detection
// =============================================================================

func TestIsSessionExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "session expired error",
			err:      status.Error(codes.FailedPrecondition, "there was an error operating on 'projects/x/sessions/y': session expired"),
			expected: true,
		},
		{
			name:     "different FailedPrecondition error",
			err:      status.Error(codes.FailedPrecondition, "some other precondition failed"),
			expected: false,
		},
		{
			name:     "NotFound error",
			err:      status.Error(codes.NotFound, "session expired"),
			expected: false,
		},
		{
			name:     "non-gRPC error",
			err:      fmt.Errorf("session expired"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSessionExpired(tt.err)
			if result != tt.expected {
				t.Errorf("isSessionExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}


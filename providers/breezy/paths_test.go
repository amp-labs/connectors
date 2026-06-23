package breezy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildVersionedPathURL(t *testing.T) {
	t.Parallel()

	u, err := buildVersionedPathURL("https://api.breezy.hr", "/companies")
	require.NoError(t, err)
	require.Equal(t, "https://api.breezy.hr/v3/companies", u.String())

	u, err = buildVersionedPathURL("https://api.breezy.hr", "/company/abc123/positions")
	require.NoError(t, err)
	require.Equal(t, "https://api.breezy.hr/v3/company/abc123/positions", u.String())
}

func TestResolveObjectPath(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"/company/abc123/positions",
		resolveObjectPath("/company/{company_id}/positions", "abc123"),
	)
}

package providers

import (
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
)

func TestResolveHTTPOptionsConnectWise(t *testing.T) {
	t.Parallel()

	got := ResolveHTTPOptions(ConnectWise, map[string]string{
		"clientId": "my-client-id",
		"region":   "na",
	})

	want := []common.HTTPOption{
		{In: common.HTTPOptionInHeader, Key: "ClientId", Value: "my-client-id"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestResolveHTTPOptionsSkipsMissingMetadata(t *testing.T) {
	t.Parallel()

	if got := ResolveHTTPOptions(ConnectWise, map[string]string{"region": "na"}); len(got) != 0 {
		t.Errorf("expected missing metadata to be skipped, got %v", got)
	}
}

func TestResolveHTTPOptionsUnregisteredProvider(t *testing.T) {
	t.Parallel()

	if got := ResolveHTTPOptions(Salesforce, map[string]string{"clientId": "x"}); got != nil {
		t.Errorf("expected nil for provider without specs, got %v", got)
	}
}

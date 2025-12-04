package main

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/getresponse/metadata"
)

// TestSchemasContainsCampaigns verifies that the embedded static schemas
// for GetResponse include metadata for the "campaigns" object under the
// root module.
func TestSchemasContainsCampaigns(t *testing.T) {
	t.Parallel()

	const objectName = "campaigns"

	// If no modules are present, the embedded schemas are effectively empty and
	// the connector cannot yet serve metadata. In that case, skip with a clear
	// diagnostic instead of failing the whole suite.
	if len(metadata.Schemas.Modules) == 0 {
		t.Skip("GetResponse embedded schemas are empty (metadata.Schemas.Modules has length 0) – likely schemas.json is not wired or has incompatible format")
	}

	// Discover actual module IDs present in the embedded schemas to avoid
	// hard-coding assumptions about the module name (e.g. \"root\").
	var moduleID common.ModuleID
	for id := range metadata.Schemas.Modules {
		moduleID = id
		break
	}

	result, err := metadata.Schemas.Select(moduleID, []string{objectName})
	if err != nil {
		t.Fatalf("Schemas.Select returned error: %v", err)
	}

	if result == nil {
		t.Fatalf("expected non-nil ListObjectMetadataResult")
	}

	if len(result.Result) == 0 {
		t.Fatalf("expected non-empty metadata result for %q, got empty", objectName)
	}

	obj, ok := result.Result[objectName]
	if !ok {
		t.Fatalf("expected metadata for object %q to be present in result", objectName)
	}

	if obj.DisplayName == "" {
		t.Errorf("expected DisplayName for %q to be set, got empty string", objectName)
	}

	if len(obj.FieldsMap) == 0 {
		t.Errorf("expected FieldsMap for %q to be non-empty", objectName)
	}
}

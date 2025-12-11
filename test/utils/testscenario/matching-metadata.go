package testscenario

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils"
)

type connectorReadMetadata interface {
	connectors.ReadConnector
	connectors.ObjectMetadataConnector
}

type readResponsePostProcessor func(before map[string]any) (after map[string]any)

// ValidateMetadataExactlyMatchesRead verifies that *all* metadata fields returned by ListObjectMetadata
// are present in the Read response for the object.
//
// If any field from the metadata list is missing in the Read output, the test fails.
// Use this when you expect Read to return *exactly* the same set of metadata fields.
//
// Use this for connectors where every possible field is always present in Read,
// so the metadata must match exactly. For connectors with optional fields,
// see ValidateMetadataContainsRead instead.
func ValidateMetadataExactlyMatchesRead(ctx context.Context, conn connectorReadMetadata, objectName string) {
	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for connector", "error", err)
	}

	slog.Info("Reading an object using all fields from ListObjectMetadata", "objectName", objectName)

	requestFields := datautils.Map[string, common.FieldMetadata](
		metadata.Result[objectName].Fields,
	).KeySet()

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     requestFields,
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	givenFields := datautils.Map[string, any](response.Data[0].Fields).KeySet()

	difference := givenFields.Diff(requestFields)
	if len(difference) != 0 {
		utils.Fail("connector read didn't match requested fields", "difference", difference)
	}

	slog.Info("==> success fields requested from ListObjectMetadata are all present in Read.")

	fmt.Println("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}

// ValidateMetadataContainsRead checks that every field present in the Read response
// is declared in the metadata returned by ListObjectMetadata for the same object.
//
// This ensures that the connector does not return any undocumented or unexpected fields:
// the metadata acts as a schema, and Read must conform to it.
//
// It is valid if the metadata includes fields that are not returned by Read
// (e.g., because they are optional or empty).
//
// The test fails if any field in the Read output is missing from the declared metadata,
// and it returns a combined error listing all such violations.
func ValidateMetadataContainsRead(
	ctx context.Context,
	conn connectorReadMetadata,
	objectName string,
	responsePostProcess readResponsePostProcessor,
) {
	metadata, err := conn.ListObjectMetadata(ctx, []string{
		objectName,
	})
	if err != nil {
		utils.Fail("error listing metadata for connector", "error", err)
	}

	slog.Info("Reading an object using all fields from ListObjectMetadata", "objectName", objectName)

	requestFields := datautils.MergeSets(
		datautils.Map[string, common.FieldMetadata](
			metadata.Result[objectName].Fields,
		).KeySet(),
		datautils.Map[string, string](
			metadata.Result[objectName].FieldsMap,
		).KeySet(),
	)

	response, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     requestFields,
	})
	if err != nil {
		utils.Fail("error reading from connector", "error", err)
	}

	if response.Rows == 0 {
		utils.Fail("expected to read at least one record", "error", err)
	}

	slog.Info("Compare object metadata against endpoint response")

	if responsePostProcess == nil {
		responsePostProcess = goutils.Identity[map[string]any]
	}

	rawData := responsePostProcess(response.Data[0].Raw)

	if mismatchErr := compareFieldsMatch(objectName, metadata, rawData); mismatchErr != nil {
		utils.Fail("schema and payload response have mismatching fields", "error", mismatchErr)
	}

	slog.Info("==> Success fields match!")

	fmt.Println("Metadata...")
	utils.DumpJSON(metadata, os.Stdout)
}

func compareFieldsMatch(objectName string, metadata *common.ListObjectMetadataResult, response map[string]any) error {
	fields := make(map[string]bool)

	for field := range response {
		fields[field] = false
	}

	mismatch := make([]error, 0)

	for fieldName := range metadata.Result[objectName].Fields {
		if _, found := fields[fieldName]; found {
			fields[fieldName] = true
		}
	}

	// every field from Read must be known to ListObjectMetadata
	for name, found := range fields {
		if !found {
			mismatch = append(mismatch, fmt.Errorf("metadata schema is missing field %v", name))
		}
	}

	return errors.Join(mismatch...)
}

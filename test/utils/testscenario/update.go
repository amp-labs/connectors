package testscenario

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

type ConnectorReadWrite interface {
	connectors.ReadConnector
	connectors.WriteConnector
}

// UpdateTestSuite controls the ValidateUpdate procedure.
type UpdateTestSuite struct {
	// ReadFields lists all fields required for read/update validation. Required.
	ReadFields datautils.StringSet

	// WaitBeforeSearch adds a delay after UPDATE if the provider needs time to reflect changes. Optional.
	WaitBeforeSearch time.Duration

	// SearchBy locates an existing record to update. If zero, the first record from Read is used.
	SearchBy Property

	// RecordIdentifierKey is the field name representing the record ID. Required.
	RecordIdentifierKey string

	// PreprocessUpdatePayload optionally adjusts the update payload based on the located record.
	PreprocessUpdatePayload func(record *common.ReadResultRow) any

	// UpdatedFields maps field names to their expected values after an update.
	UpdatedFields map[string]string

	// ValidateUpdatedFields optionally verifies that updates took effect correctly.
	ValidateUpdatedFields func(record map[string]any)

	// RestoreOriginalFields reverts updated fields to their pre-update values after validation.
	RestoreOriginalFields bool
}

// ValidateUpdate is a test scenario for objects that support update only.
//
// Flow:
// 1. Read and locate an existing object.
// 2. Update the object using the "UP" payload.
// 3. Read again and verify updates took effect.
// 4. Optionally restore the original field values.
func ValidateUpdate[UP any](
	ctx context.Context, conn ConnectorReadWrite, objectName string,
	updatePayload UP, suite UpdateTestSuite,
) {
	fmt.Println("> TEST Update", objectName)

	fmt.Println("Reading", objectName)

	res, err := readObjects(ctx, conn, objectName, suite.ReadFields, suite.SearchBy.Since)
	failOnError(err)

	object, err := locateExistingRecord(res, suite.SearchBy)
	failOnError(err)

	objectID := object.getRecordIdentifierValue(suite.RecordIdentifierKey)
	fmt.Println("Object record identifier is", objectID)

	preprocessPayloadFunc := suite.PreprocessUpdatePayload
	if preprocessPayloadFunc == nil {
		preprocessPayloadFunc = func(*common.ReadResultRow) any { return updatePayload }
	}

	payload := preprocessPayloadFunc(&object.ReadResultRow)

	payloadMap, err := common.RecordDataToMap(payload)
	failOnError(err)

	originalValues := originalFieldValues(object.Fields, payloadMap)

	fmt.Println("Updating", objectName)

	_, err = updateObject(ctx, conn, objectName, objectID, &payloadMap)
	failOnError(err)

	fmt.Println("Validate object has changed accordingly")

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	validateUpdatedFieldsFunc := suite.ValidateUpdatedFields
	if validateUpdatedFieldsFunc == nil {
		validateUpdatedFieldsFunc = func(record map[string]any) {
			for k, v := range suite.UpdatedFields {
				if !mockutils.DoesObjectCorrespondToString(record[k], v) {
					utils.Fail("error updated properties do not match", k, v, record[k])
				}
			}
		}
	}

	res, err = readObjects(ctx, conn, objectName, suite.ReadFields, suite.SearchBy.Since)
	failOnError(err)

	object, err = searchObjectRecord(res, suite.RecordIdentifierKey, objectID)
	failOnError(err)
	validateUpdatedFieldsFunc(object.Fields)

	if suite.RestoreOriginalFields && len(originalValues) > 0 {
		fmt.Println("Restoring original field values")

		_, err = updateObject(ctx, conn, objectName, objectID, &originalValues)
		failOnError(err)
	}

	fmt.Println("> Successful test completion")
}

func locateExistingRecord(res *common.ReadResult, searchBy Property) (*objectRecord, error) {
	if !searchBy.isZero() {
		fmt.Println("Finding", "record by", searchBy)

		return searchObjectRecord(res, searchBy.Key, searchBy.Value)
	}

	fmt.Println("Using first available record")

	if len(res.Data) == 0 {
		return nil, errors.New("no records found to update")
	}

	return &objectRecord{res.Data[0]}, nil
}

func originalFieldValues(fields map[string]any, payload map[string]any) map[string]any {
	original := make(map[string]any, len(payload))

	for key := range payload {
		if value, ok := fields[key]; ok {
			original[key] = value
		}
	}

	return original
}

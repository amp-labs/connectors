package testscenario

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

// CRDTestSuite controls the ValidateCreateDelete procedure.
type CRDTestSuite struct {
	// ReadFields lists all fields required for read validation. Required.
	ReadFields datautils.StringSet

	// WaitBeforeSearch adds a delay after CREATE if the provider needs time to reflect changes. Optional.
	WaitBeforeSearch time.Duration

	// SearchBy specifies a unique property used to locate the record in the list. Required.
	// This lookup is performed once, and the record ID is saved and reused throughout the test scenario.
	SearchBy Property

	// RecordIdentifierKey is the field name representing the record ID.
	RecordIdentifierKey string
}

// ValidateCreateDelete is a comprehensive test scenario utilizing Read/Write/Delete connector operations.
//
// Flow:
// 1. Create an object using the "CP" payload.
// 2. Read and locate the object using test-defined criteria.
// 3. Delete the object at the end.
func ValidateCreateDelete[CP any](ctx context.Context, conn ConnectorCRUD, objectName string,
	createPayload CP, suite CRDTestSuite,
) {
	fmt.Println("> TEST Create/Delete", objectName)
	fmt.Println("Creating", objectName)
	createObject(ctx, conn, objectName, &createPayload)

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	fmt.Println("Reading", objectName)

	res := readObjects(ctx, conn, objectName, suite.ReadFields, suite.SearchBy.Since)

	fmt.Println("Finding recently created", objectName)

	search := suite.SearchBy
	object := searchObjectRecord(res, search.Key, search.Value)
	objectID := object.getRecordIdentifierValue(suite.RecordIdentifierKey)

	fmt.Println("Object record identifier is", objectID)

	fmt.Println("Removing this", objectName)
	removeObject(ctx, conn, objectName, objectID)
	fmt.Println("> Successful test completion")
}

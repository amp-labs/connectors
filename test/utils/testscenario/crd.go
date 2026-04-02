package testscenario

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
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
	_, objectID, err := createAndFindRecord(ctx, conn, objectName, createPayload, suite)
	failOnError(err)

	fmt.Println("Object record identifier is", objectID)

	fmt.Println("Removing this", objectName)
	err = removeObject(ctx, conn, objectName, objectID)
	failOnError(err)
	fmt.Println("> Successful test completion")
}

type RecordCreationRecipe CRDTestSuite

// SetupRecord prepares a single record of the given object type in the provider,
// returning both the read representation of the record and a cleanup function.
// Use it to create temporary records used by your test case that should be removed after completion.
//
// Flow:
// 1. Create an object using the "CP" payload.
// 2. Read and locate the object using the criteria defined in creationRecipe (RecordCreationRecipe).
// 3. Return:
//   - The read result row (common.ReadResultRow) of the created object.
//   - A cleanup function that, when called, deletes the object by its ID.
//
// Usage pattern:
//
//	record, cleanup := testscenario.SetupRecord(ctx, conn, "Users", CreatePayload{...}, suite)
//	defer cleanup()
//	userID := record.Fields["id"]
func SetupRecord[CP any](ctx context.Context, conn ConnectorCRUD, objectName string,
	createPayload CP, creationRecipe RecordCreationRecipe,
) (record *common.ReadResultRow, cleanup func(), errOut error) {
	fmt.Printf("[SETUP] Creating a record for object '%v'\n", objectName)
	object, objectID, err := createAndFindRecord(ctx, conn, objectName, createPayload, CRDTestSuite(creationRecipe))
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Object %v(%v) is ready\n", objectName, objectID)

	cleanup = func() {
		removeErr := removeObject(ctx, conn, objectName, objectID)
		if removeErr != nil {
			fmt.Printf("[CLEANUP] Record removal FAILED! Please, remove object yourself.\n"+
				"\tObject(%v)\n\tID(%v)\n\tReason: %v\n", objectName, objectID, removeErr)
			return
		}

		fmt.Printf("[CLEANUP] Successful removal\n"+
			"\tObject(%v)\n\tID(%v)\n", objectName, objectID)
	}

	return &object.ReadResultRow, cleanup, nil
}

func createAndFindRecord[CP any](
	ctx context.Context, conn ConnectorCRUD, objectName string, createPayload CP, suite CRDTestSuite,
) (*objectRecord, string, error) {
	_, err := createObject(ctx, conn, objectName, &createPayload)
	if err != nil {
		return nil, "", err
	}

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	fmt.Println("Reading", objectName)

	res, err := readObjects(ctx, conn, objectName, suite.ReadFields, suite.SearchBy.Since)
	if err != nil {
		return nil, "", err
	}

	fmt.Println("Finding recently created", objectName)

	search := suite.SearchBy
	object, err := searchObjectRecord(res, search.Key, search.Value)
	if err != nil {
		return nil, "", err
	}

	objectID := object.getRecordIdentifierValue(suite.RecordIdentifierKey)

	return object, objectID, nil
}

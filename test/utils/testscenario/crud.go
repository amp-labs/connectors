package testscenario

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

type ConnectorCRUD interface {
	connectors.ReadConnector
	connectors.WriteConnector
	connectors.DeleteConnector
}

// CRUDTestSuite controls the ValidateCreateUpdateDelete procedure.
type CRUDTestSuite struct {
	// ReadFields lists all fields required for read/update validation. Required.
	ReadFields datautils.StringSet

	// WaitBeforeSearch adds a delay after CREATE/UPDATE if the provider needs time to reflect changes. Optional.
	WaitBeforeSearch time.Duration

	// SearchBy specifies a unique property used to locate the record in the list. Required.
	// This lookup is performed once, and the record ID is saved and reused throughout the test scenario.
	SearchBy Property

	// RecordIdentifierKey is the field name representing the record ID.
	RecordIdentifierKey string

	// UpdatedFields maps field names to expected values for verifying that the update took effect.
	UpdatedFields map[string]string
}

type Property struct {
	Key   string
	Value string
}

// ValidateCreateUpdateDelete is a comprehensive test scenario utilizing Read/Write/Delete connector operations.
//
// Flow:
// 1. Create an object using the "CP" payload.
// 2. Read and locate the object using test-defined criteria.
// 3. Update the object using the "UP" payload.
// 4. Read again and verify updates took effect.
// 5. Delete the object at the end.
func ValidateCreateUpdateDelete[CP, UP any](ctx context.Context, conn ConnectorCRUD, objectName string,
	createPayload CP, updatePayload UP, suite CRUDTestSuite,
) {
	fmt.Println("> TEST Create/Update/Delete", objectName)
	fmt.Println("Creating", objectName)
	createObject(ctx, conn, objectName, &createPayload)

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	fmt.Println("Reading", objectName)

	res := readObjects(ctx, conn, objectName, suite.ReadFields)

	fmt.Println("Finding recently created", objectName)

	search := suite.SearchBy
	object := searchObject(res, search.Key, search.Value)
	objectID := getRecordIdentifierValue(object, suite.RecordIdentifierKey)

	fmt.Println("Updating some object properties")
	updateObject(ctx, conn, objectName, objectID, &updatePayload)
	fmt.Println("Validate object has changed accordingly")

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	res = readObjects(ctx, conn, objectName, suite.ReadFields)

	object = searchObject(res, suite.RecordIdentifierKey, objectID)
	for k, v := range suite.UpdatedFields {
		if !mockutils.DoesObjectCorrespondToString(object[k], v) {
			utils.Fail("error updated properties do not match", k, v, object[k])
		}
	}

	fmt.Println("Removing this", objectName)
	removeObject(ctx, conn, objectName, objectID)
	fmt.Println("> Successful test completion")
}

func getRecordIdentifierValue(object map[string]any, key string) string {
	switch id := object[key].(type) {
	case string:
		return id
	case float64:
		return strconv.FormatFloat(id, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(id, 10)
	case int:
		return strconv.Itoa(id)
	case uint64:
		return strconv.FormatUint(id, 10)
	default:
		return fmt.Sprintf("%v", id)
	}
}

func createObject[CP any](ctx context.Context, conn ConnectorCRUD, objectName string, payload *CP) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error creating an object", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create an object")
	}
}

func readObjects(ctx context.Context, conn ConnectorCRUD, objectName string, fields datautils.StringSet) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     fields,
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	return res
}

func searchObject(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding object in a list")

	return nil
}

func updateObject[UP any](ctx context.Context, conn ConnectorCRUD, objectName string, objectID string, payload *UP) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   objectID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error updating object", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update an object")
	}
}

func removeObject(ctx context.Context, conn ConnectorCRUD, objectName string, objectID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   objectID,
	})
	if err != nil {
		utils.Fail("error deleting for provider", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove object")
	}
}

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

	// SearchBy specifies a unique property used to locate the newly created record.
	// If set, the test searches the read response for a matching record and extracts its ID.
	// If not set, the record ID from the create response is used instead.
	SearchBy Property

	// RecordIdentifierKey is the field name representing the record ID.
	RecordIdentifierKey string

	// PreprocessUpdatePayload optionally adjusts the update payload based on
	// data returned by the create operation.
	PreprocessUpdatePayload func(createResult *common.WriteResult, updatePayload any)

	// UpdatedFields maps field names to their expected values after an update.
	// For simple flat fields, use this. For complex or nested structures,
	// prefer ValidateUpdatedFields.
	UpdatedFields map[string]string

	// ValidateUpdatedFields optionally verifies that updates took effect correctly.
	// Use this when UpdatedFields is insufficient — for example, when validating nested arrays or objects.
	ValidateUpdatedFields func(record map[string]any)
}

type Property struct {
	Key   string
	Value string
}

func (p Property) isZero() bool {
	return p.Key == "" && p.Value == ""
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

	// CREATE
	fmt.Println("Creating", objectName)
	createResult := createObject(ctx, conn, objectName, &createPayload)

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	// READ
	fmt.Println("Reading", objectName)
	res := readObjects(ctx, conn, objectName, suite.ReadFields)

	// SEARCH
	fmt.Println("Finding recently created", objectName)

	objectID := createResult.RecordId

	if !suite.SearchBy.isZero() {
		search := suite.SearchBy
		object := searchObject(res, search.Key, search.Value)
		objectID = getRecordIdentifierValue(object, suite.RecordIdentifierKey)
	}

	fmt.Println("Object record identifier is", objectID)

	// PREPROCESS UPDATE PAYLOAD
	preprocessPayloadFunc := suite.PreprocessUpdatePayload
	if preprocessPayloadFunc == nil {
		// By default, update payload doesn't depend on data from create response.
		preprocessPayloadFunc = func(*common.WriteResult, any) {}
	}

	preprocessPayloadFunc(createResult, &updatePayload)

	// UPDATE
	fmt.Println("Updating some object properties")
	updateObject(ctx, conn, objectName, objectID, &updatePayload)
	fmt.Println("Validate object has changed accordingly")

	if suite.WaitBeforeSearch != 0 {
		fmt.Println("... waiting")
		time.Sleep(suite.WaitBeforeSearch)
	}

	// VALIDATE UPDATE
	validateUpdatedFieldsFunc := suite.ValidateUpdatedFields
	if validateUpdatedFieldsFunc == nil {
		// By default, compare each field to string counterpart.
		// Complicated responses with nested objects or arrays require custom func.
		validateUpdatedFieldsFunc = func(record map[string]any) {
			for k, v := range suite.UpdatedFields {
				if !mockutils.DoesObjectCorrespondToString(record[k], v) {
					utils.Fail("error updated properties do not match", k, v, record[k])
				}
			}
		}
	}

	res = readObjects(ctx, conn, objectName, suite.ReadFields)
	object := searchObject(res, suite.RecordIdentifierKey, objectID)
	validateUpdatedFieldsFunc(object)

	// DELETE
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

func createObject[CP any](ctx context.Context, conn ConnectorCRUD, objectName string, payload *CP) *common.WriteResult {
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

	return res
}

func readObjects(ctx context.Context, conn ConnectorCRUD, objectName string, fields datautils.StringSet) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     fields,
	})
	if err != nil {
		utils.Fail("error reading from provider", "error", err)
	}

	// Paginate
	for len(res.NextPage) > 0 {
		params := common.ReadParams{
			ObjectName: objectName,
			Fields:     fields,
			NextPage:   res.NextPage,
		}

		nextRes, err := conn.Read(ctx, params)
		if err != nil {
			utils.Fail("error reading next page from provider", "error", err)
		}

		res.Data = append(res.Data, nextRes.Data...)
		res.NextPage = nextRes.NextPage
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

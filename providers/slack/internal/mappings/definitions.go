package mappings

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

type Object struct {
	readListInfo    *ReadListInfo    // Used by Read and ListObjectMetadata
	readItemInfo    *ReadItemInfo    // Used by Batch Read
	writeCreateInfo *WriteCreateInfo // Used by Write
	writeUpdateInfo *WriteUpdateInfo // Used by Write
	deleteInfo      *DeleteInfo      // Used by Delete
}

type ReadListInfo struct {
	// Href is a full name of slack operation.
	Href string
	// Method is http method used to make an API call.
	Method string
	// ResponseField holds the array of object records.
	ResponseField string
	// NestedResponseIdField specifies the location of ResponseIdField from root instead of relative to ResponseField.
	NestedResponseIdField []string
	// ResponseIdField is id location of single record.
	ResponseIdField string
	// TimeFilterField used by connector side filtering.
	TimeFilterField string
	// SinceQP is since query param.
	SinceQP string
	// UntilQP is until query param.
	UntilQP string
}

type ReadItemInfo struct {
	Href            string
	Method          string
	RequestIdField  string
	ResponseField   string
	ResponseIdField string
}

type WriteCreateInfo struct {
	Href            string
	ResponseField   string
	ResponseIdField string
}

type WriteUpdateInfo struct {
	Href            string
	RequestIdField  string
	ResponseField   string
	ResponseIdField string
}

type DeleteInfo struct {
	Href           string
	Method         string
	RequestIdField string
}

func GetReadListInfo(provider providers.Provider, objectName string) (ReadListInfo, error) {
	object, err := getObject(provider, objectName)
	if err != nil {
		return ReadListInfo{}, err
	}

	if object.readListInfo == nil {
		return ReadListInfo{}, common.ErrOperationNotSupportedForObject
	}

	return *object.readListInfo, nil
}

func GetReadItemInfo(provider providers.Provider, objectName string) (ReadItemInfo, error) {
	object, err := getObject(provider, objectName)
	if err != nil {
		return ReadItemInfo{}, err
	}

	if object.readItemInfo == nil {
		return ReadItemInfo{}, common.ErrOperationNotSupportedForObject
	}

	return *object.readItemInfo, nil
}

func GetWriteCreateInfo(provider providers.Provider, objectName string) (WriteCreateInfo, error) {
	object, err := getObject(provider, objectName)
	if err != nil {
		return WriteCreateInfo{}, err
	}

	if object.writeCreateInfo == nil {
		return WriteCreateInfo{}, common.ErrOperationNotSupportedForObject
	}

	return *object.writeCreateInfo, nil
}

func GetWriteUpdateInfo(provider providers.Provider, objectName string) (WriteUpdateInfo, error) {
	object, err := getObject(provider, objectName)
	if err != nil {
		return WriteUpdateInfo{}, err
	}

	if object.writeUpdateInfo == nil {
		return WriteUpdateInfo{}, common.ErrOperationNotSupportedForObject
	}

	return *object.writeUpdateInfo, nil
}

func GetDeleteInfo(provider providers.Provider, objectName string) (DeleteInfo, error) {
	object, err := getObject(provider, objectName)
	if err != nil {
		return DeleteInfo{}, err
	}

	if object.deleteInfo == nil {
		return DeleteInfo{}, common.ErrOperationNotSupportedForObject
	}

	return *object.deleteInfo, nil
}

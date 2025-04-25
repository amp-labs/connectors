package awsic

import (
	"encoding/json"
	"errors"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
)

var ErrCommandNotFound = errors.New("AWS command was not found for the object")

// TODO construct payload depending on the Service+Operation
// TODO Service is implied from the object (otherwise connector will have many modules, need investigation)
type serviceCommand struct {
	ServiceDomain  string
	ServiceName    string
	Command        string
	PayloadBuilder func(values map[string]string) []byte
}

var readCommands = datautils.Map[string, serviceCommand]{
	"Instances": {
		ServiceDomain: "sso",
		ServiceName:   "SWBExternalService",
		Command:       "ListInstances",
		PayloadBuilder: func(values map[string]string) []byte {
			return servicePayload{
				MaxResults:  10,
				InstanceArn: goutils.Pointer(values["InstanceArn"]),
			}.toBytes()
		},
	},
	// TODO this is not working
	// Error from the provider:
	// --> An error occurred (ValidationException) when calling the ListPermissionSets operation: The operation is not supported for this Identity Center instance
	"PermissionSets": {
		ServiceDomain: "sso",
		ServiceName:   "SWBExternalService",
		Command:       "ListPermissionSets",
		PayloadBuilder: func(values map[string]string) []byte {
			return servicePayload{
				MaxResults: 10,
			}.toBytes()
		},
	},
	"Groups": {
		ServiceDomain: "identitystore",
		ServiceName:   "AWSIdentityStore",
		Command:       "ListGroups",
		PayloadBuilder: func(values map[string]string) []byte {
			return servicePayload{
				MaxResults:      10,
				IdentityStoreId: goutils.Pointer(values["IdentityStoreID"]),
			}.toBytes()
		},
	},
	"Users": {
		ServiceDomain: "identitystore",
		ServiceName:   "AWSIdentityStore",
		Command:       "ListUsers",
		PayloadBuilder: func(values map[string]string) []byte {
			return servicePayload{
				MaxResults:      10,
				IdentityStoreId: goutils.Pointer(values["IdentityStoreID"]),
			}.toBytes()
		},
	},
}

type servicePayload struct {
	MaxResults      int     `json:"MaxResults"`
	IdentityStoreId *string `json:"IdentityStoreId,omitempty"`
	InstanceArn     *string `json:"InstanceArn,omitempty"`
}

func (p servicePayload) toBytes() []byte {
	data, _ := json.Marshal(p)

	return data
}

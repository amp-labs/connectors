// nolint:gochecknoglobals
package ssoadmin

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	ServiceName        = "SWBExternalService"
	ServiceDomain      = "sso"
	ServiceSigningName = "sso"
)

var ReadObjectCommands = datautils.Map[string, string]{
	"Instances":                       "ListInstances",
	"Applications":                    "ListApplications",
	"ApplicationProviders":            "ListApplicationProviders",
	"AccountAssignmentCreationStatus": "ListAccountAssignmentCreationStatus",
	"AccountAssignmentDeletionStatus": "ListAccountAssignmentDeletionStatus",
	"PermissionSetProvisioningStatus": "ListPermissionSetProvisioningStatus",
	"TrustedTokenIssuers":             "ListTrustedTokenIssuers",
}

package ssoadmin

import (
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

const (
	ServiceName        = "SWBExternalService"
	ServiceDomain      = "sso"
	ServiceSigningName = "sso"
)

var Registry = core.Registry{ // nolint:gochecknoglobals
	"AccountAssignmentCreationStatus": {
		Commands: core.ObjectCommands{
			Read:   "ListAccountAssignmentCreationStatus",
			Create: "",
			Update: "",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"AccountAssignmentDeletionStatus": {
		Commands: core.ObjectCommands{
			Read:   "ListAccountAssignmentDeletionStatus",
			Create: "",
			Update: "",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"AccountAssignments": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "CreateAccountAssignment",
			Update: "",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"ApplicationAccessScopes": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "PutApplicationAccessScope",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "ApplicationArn",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"ApplicationAssignmentConfigurations": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "PutApplicationAssignmentConfiguration",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "ApplicationArn",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"ApplicationAssignments": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "CreateApplicationAssignment",
			Update: "",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"ApplicationAuthenticationMethods": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "PutApplicationAuthenticationMethod",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "ApplicationArn",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"ApplicationGrants": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "PutApplicationGrant",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "ApplicationArn",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"ApplicationProviders": {
		Commands: core.ObjectCommands{
			Read:   "ListApplicationProviders",
			Create: "",
			Update: "",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"Applications": {
		Commands: core.ObjectCommands{
			Read:   "ListApplications",
			Create: "CreateApplication",
			Update: "UpdateApplication",
			Delete: "DeleteApplication",
		},
		InputRecordID: core.InputRecordID{
			Update: "ApplicationArn",
			Delete: "ApplicationArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("ApplicationArn"),
			Update: nil,
		},
	},
	"InlinePolicyFromPermissionSets": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "",
			Delete: "DeleteInlinePolicyFromPermissionSet",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "PermissionSetArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"InlinePolicyToPermissionSets": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "PutInlinePolicyToPermissionSet",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "PermissionSetArn",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"InstanceAccessControlAttributeConfigurations": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "CreateInstanceAccessControlAttributeConfiguration",
			Update: "", // unclear what record id should be
			Delete: "DeleteInstanceAccessControlAttributeConfiguration",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "InstanceArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"Instances": {
		Commands: core.ObjectCommands{
			Read:   "ListInstances",
			Create: "CreateInstance",
			Update: "UpdateInstance",
			Delete: "DeleteInstance",
		},
		InputRecordID: core.InputRecordID{
			Update: "InstanceArn",
			Delete: "InstanceArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("InstanceArn"),
			Update: nil,
		},
	},
	"PermissionSetProvisioningStatus": {
		Commands: core.ObjectCommands{
			Read:   "ListPermissionSetProvisioningStatus",
			Create: "",
			Update: "",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"PermissionSets": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "CreatePermissionSet",
			Update: "UpdatePermissionSet",
			Delete: "DeletePermissionSet",
		},
		InputRecordID: core.InputRecordID{
			Update: "PermissionSetArn",
			Delete: "PermissionSetArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("PermissionSetArn", "PermissionSet"),
			Update: nil,
		},
	},
	"PermissionsBoundaryFromPermissionSets": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "",
			Delete: "DeletePermissionsBoundaryFromPermissionSet",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "PermissionSetArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"PermissionsBoundaryToPermissionSets": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "",
			Update: "PutPermissionsBoundaryToPermissionSet",
			Delete: "",
		},
		InputRecordID: core.InputRecordID{
			Update: "PermissionSetArn",
			Delete: "",
		},
		OutputRecordID: core.OutputRecordID{
			Create: nil,
			Update: nil,
		},
	},
	"TrustedTokenIssuers": {
		Commands: core.ObjectCommands{
			Read:   "ListTrustedTokenIssuers",
			Create: "CreateTrustedTokenIssuer",
			Update: "UpdateTrustedTokenIssuer",
			Delete: "DeleteTrustedTokenIssuer",
		},
		InputRecordID: core.InputRecordID{
			Update: "TrustedTokenIssuerArn",
			Delete: "TrustedTokenIssuerArn",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("TrustedTokenIssuerArn"),
			Update: nil,
		},
	},
}

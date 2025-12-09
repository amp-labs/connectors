package identitystore

import (
	"github.com/amp-labs/connectors/providers/aws/internal/core"
)

const (
	ServiceName        = "AWSIdentityStore"
	ServiceDomain      = "identitystore"
	ServiceSigningName = "identitystore"
)

var Registry = core.Registry{ // nolint:gochecknoglobals
	"Users": {
		Commands: core.ObjectCommands{
			Read:   "ListUsers",
			Create: "CreateUser",
			Update: "UpdateUser",
			Delete: "DeleteUser",
		},
		InputRecordID: core.InputRecordID{
			Update: "UserId",
			Delete: "UserId",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("UserId"),
			Update: nil,
		},
	},
	"Groups": {
		Commands: core.ObjectCommands{
			Read:   "ListGroups",
			Create: "CreateGroup",
			Update: "UpdateGroup",
			Delete: "DeleteGroup",
		},
		InputRecordID: core.InputRecordID{
			Update: "GroupId",
			Delete: "GroupId",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("GroupId"),
			Update: nil,
		},
	},
	"GroupMemberships": {
		Commands: core.ObjectCommands{
			Read:   "",
			Create: "CreateGroupMembership",
			Update: "",
			Delete: "DeleteGroupMembership",
		},
		InputRecordID: core.InputRecordID{
			Update: "",
			Delete: "MembershipId",
		},
		OutputRecordID: core.OutputRecordID{
			Create: core.NewRecordLocation("MembershipId"),
			Update: nil,
		},
	},
}

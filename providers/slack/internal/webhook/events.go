package webhook

import "github.com/amp-labs/connectors/common"

const (
	typeCreate = common.SubscriptionEventTypeCreate
	typeUpdate = common.SubscriptionEventTypeUpdate
	typeDelete = common.SubscriptionEventTypeDelete
	typeOther  = common.SubscriptionEventTypeOther
)

type eventDescription struct {
	ObjectName string
	Type       common.SubscriptionEventType
}

// nolint:lll,gochecknoglobals
var eventNameToEventDescription = map[string]eventDescription{
	"bot_added":               {ObjectName: "bots", Type: typeCreate},          // https://docs.slack.dev/reference/events/bot_added
	"bot_changed":             {ObjectName: "bots", Type: typeUpdate},          // https://docs.slack.dev/reference/events/bot_changed
	"call_rejected":           {ObjectName: "calls", Type: typeOther},          // https://docs.slack.dev/reference/events/call_rejected
	"channel_archive":         {ObjectName: "conversations", Type: typeOther},  // https://docs.slack.dev/reference/events/channel_archive
	"channel_created":         {ObjectName: "conversations", Type: typeCreate}, // https://docs.slack.dev/reference/events/channel_created
	"channel_deleted":         {ObjectName: "conversations", Type: typeDelete}, // https://docs.slack.dev/reference/events/channel_deleted
	"channel_history_changed": {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/channel_history_changed
	"channel_id_changed":      {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/channel_id_changed
	"channel_joined":          {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/channel_joined
	"channel_left":            {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/channel_left
	"channel_rename":          {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/channel_rename
	"channel_unarchive":       {ObjectName: "conversations", Type: typeOther},  // https://docs.slack.dev/reference/events/channel_unarchive
	"file_change":             {ObjectName: "files", Type: typeUpdate},         // https://docs.slack.dev/reference/events/file_change
	"file_created":            {ObjectName: "files", Type: typeCreate},         // https://docs.slack.dev/reference/events/file_created
	"file_deleted":            {ObjectName: "files", Type: typeDelete},         // https://docs.slack.dev/reference/events/file_deleted
	"group_archive":           {ObjectName: "conversations", Type: typeOther},  // https://docs.slack.dev/reference/events/group_archive
	"group_deleted":           {ObjectName: "conversations", Type: typeDelete}, // https://docs.slack.dev/reference/events/group_deleted
	"group_joined":            {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/group_joined
	"group_left":              {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/group_left
	"group_rename":            {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/group_rename
	"group_unarchive":         {ObjectName: "conversations", Type: typeOther},  // https://docs.slack.dev/reference/events/group_unarchive
	"im_created":              {ObjectName: "conversations", Type: typeCreate}, // https://docs.slack.dev/reference/events/im_created
	"im_history_changed":      {ObjectName: "conversations", Type: typeCreate}, // https://docs.slack.dev/reference/events/im_history_changed
	"member_joined_channel":   {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/member_joined_channel
	"member_left_channel":     {ObjectName: "conversations", Type: typeUpdate}, // https://docs.slack.dev/reference/events/member_left_channel
	"user_change":             {ObjectName: "users", Type: typeUpdate},         // https://docs.slack.dev/reference/events/user_change
	"user_profile_changed":    {ObjectName: "users", Type: typeUpdate},         // https://docs.slack.dev/reference/events/user_profile_changed
	"user_status_changed":     {ObjectName: "users", Type: typeUpdate},         // https://docs.slack.dev/reference/events/user_status_changed/
}

package zoho

import (
	"errors"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	OperationCreate = "create"
	OperationEdit   = "edit"
	OperationDelete = "delete"
	OperationAll    = "all"
	deskLimit       = "50"
	deskLimitInt    = 50
	createdTimeKey  = "createdTime"
	modifiedTimeKey = "modifiedTime"

	maxWatchFields = 10

	ResultStatusSuccess = "SUCCESS"
	defaultDuration     = 7 * 24 * time.Hour // 1 week and this is max duration for subscription
)

var (
	errInvalidRequestType    = errors.New("invalid request type")
	errMissingParams         = errors.New("missing required parameters")
	errWatchFieldsAll        = errors.New("watch fields all is not supported")
	errTooManyWatchFields    = errors.New("too many watch fields")
	errSubscriptionFailed    = errors.New("subscription failed")
	errNoSubscriptionCreated = errors.New("no subscription created")
	errUnsupportedEventType  = errors.New("unsupported event type")
	errFieldNotFound         = errors.New("field not found")
	errObjectNameNotFound    = errors.New("object name not found")
	errInvalidModuleEvent    = errors.New("invalid module event")
	//nolint:revive
	errInconsistentChannelIdsMismatch = errors.New("all events must have the same channel id")
	errChannelIdMismatch              = errors.New("channel id does not match provided unique ref")
	errInvalidDuration                = errors.New("duration cannot be greater than 1 week")
	errModuleNameNotString            = errors.New("module_name is not a string")
	errAPINameNotString               = errors.New("api_name is not a string")
	errIDNotString                    = errors.New("id is not a string")
	errFieldIDNotString               = errors.New("field id is not a string")
)

var (
	endpointsWithModifiedAfterParam = datautils.NewSet("im/sessions", "im/cannedMessages")      //nolint: gochecknoglobals
	objectsSortableByCreatedTime    = datautils.NewStringSet("contacts", "tickets", "accounts", //nolint: gochecknoglobals
		"tasks", "calls", "events", "ticketTags")
	objectsSortablebyModifiedTime = datautils.NewStringSet( //nolint: gochecknoglobals
		"articles", "communityTopics", "contracts", "groups", "labels", "users", "products")
)

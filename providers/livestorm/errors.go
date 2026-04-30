package livestorm

import "errors"

var (
	// ErrSessionIDRequired is returned when reading session_chat_messages without a session id in ReadParams.Filter.
	ErrSessionIDRequired = errors.New("read session_chat_messages requires a non-empty ReadParams.Filter (session id)")

	// ErrJobIDRequired is returned when reading jobs without a job id in ReadParams.Filter.
	ErrJobIDRequired = errors.New("read jobs requires a non-empty ReadParams.Filter (job id)")
)

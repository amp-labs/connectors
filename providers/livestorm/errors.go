package livestorm

import "errors"

// ErrSessionIDRequired is returned when reading session_chat_messages without a session id in ReadParams.Filter.
var ErrSessionIDRequired = errors.New("read session_chat_messages requires a non-empty ReadParams.Filter (session id)")

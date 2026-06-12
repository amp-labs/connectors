package fathom

import "errors"

var (
	ErrUnexpectedRecordingIDType = errors.New("unexpected recording_id type")
	ErrUnexpectedSummaryType     = errors.New("unexpected summary type")
)

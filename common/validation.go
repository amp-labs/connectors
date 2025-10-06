package common

import "errors"

var (
	// ErrMissingObjects is returned when no objects are provided in the request.
	ErrMissingObjects = errors.New("no objects provided")

	// ErrMissingRecordID is returned when resource id is missing in the request.
	ErrMissingRecordID = errors.New("no object ID provided")

	// ErrMissingRecordData is returned when write data is missing in the request.
	ErrMissingRecordData = errors.New("no data provided")

	// ErrMissingFields is returned when no fields are provided for reading.
	ErrMissingFields = errors.New("no fields provided in ReadParams")

	// ErrSinceUntilChronOrder is returned when the 'since' timestamp is after the 'until' timestamp in ReadParams.
	ErrSinceUntilChronOrder = errors.New("since cannot come after until")
)

func (p ReadParams) ValidateParams(withRequiredFields bool) error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if withRequiredFields && len(p.Fields) == 0 {
		return ErrMissingFields
	}

	// If both 'since' and 'until' are set, ensure correct chronological order.
	if !p.Since.IsZero() && !p.Until.IsZero() {
		// Until must be after since, otherwise error.
		if p.Since.After(p.Until) {
			return ErrSinceUntilChronOrder
		}
	}

	return nil
}

func (p WriteParams) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if p.RecordData == nil {
		return ErrMissingRecordData
	}

	return nil
}

func (p DeleteParams) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if len(p.RecordId) == 0 {
		return ErrMissingRecordID
	}

	return nil
}

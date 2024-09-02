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
)

func (p ReadParams) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if len(p.Fields) == 0 {
		return ErrMissingFields
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

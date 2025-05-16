package core

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

type ObjectProperties struct {
	Commands       ObjectCommands
	InputRecordID  InputRecordID
	OutputRecordID OutputRecordID
}

type ObjectCommands struct {
	Read   string
	Create string
	Update string
	Delete string
}

type InputRecordID struct {
	Update string
	Delete string
}

type OutputRecordID struct {
	Create *RecordLocation
	Update *RecordLocation
}

func (p ObjectProperties) CanRead() bool {
	return len(p.Commands.Read) != 0
}

func (p ObjectProperties) CanCreate() bool {
	return len(p.Commands.Create) != 0
}

func (p ObjectProperties) CanUpdate() bool {
	return len(p.Commands.Update) != 0
}

func (p ObjectProperties) CanDelete() bool {
	return len(p.Commands.Delete) != 0
}

type Registry datautils.Map[string, ObjectProperties]

func (r Registry) Has(objectName string) bool {
	return datautils.Map[string, ObjectProperties](r).Has(objectName)
}

func (r Registry) GetReadObjects() datautils.StringSet {
	result := datautils.NewStringSet()

	for objectName, props := range r {
		if props.CanRead() {
			result.AddOne(objectName)
		}
	}

	return result
}

func (r Registry) GetWriteObjects() datautils.StringSet {
	result := datautils.NewStringSet()

	for objectName, props := range r {
		if props.CanCreate() || props.CanUpdate() {
			result.AddOne(objectName)
		}
	}

	return result
}

func (r Registry) GetDeleteObjects() datautils.StringSet {
	result := datautils.NewStringSet()

	for objectName, props := range r {
		if props.CanDelete() {
			result.AddOne(objectName)
		}
	}

	return result
}

type RecordLocation struct {
	ID   string
	Zoom []string
}

func NewRecordLocation(id string, zoom ...string) *RecordLocation {
	return &RecordLocation{
		ID:   id,
		Zoom: zoom,
	}
}

func (l *RecordLocation) Extract(node *ajson.Node, defaultValue string) string {
	if l == nil {
		return ""
	}

	result, err := jsonquery.New(node, l.Zoom...).TextWithDefault(l.ID, defaultValue)
	if err != nil {
		return ""
	}

	return result
}

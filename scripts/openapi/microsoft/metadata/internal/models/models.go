package models

import (
	"errors"
	"maps"
	"strings"

	"github.com/amp-labs/connectors/internal/metadatadef"
)

var ErrInvalidParentRef = errors.New("referenced parent type doesn't exist for schema")

// EntitySet organised by Entity.Name.
type EntitySet map[string]*Entity

func NewEntitySet() EntitySet {
	return make(map[string]*Entity)
}

func (s EntitySet) GetNames() []string {
	names := make([]string, len(s))
	i := 0

	for name := range s {
		names[i] = name
		i++
	}

	return names
}

func (s EntitySet) GetOrCreate(name, parentName string, isAbstract bool) *Entity {
	if _, ok := s[name]; !ok {
		s[name] = &Entity{
			Name:       name,
			properties: make(map[string]metadatadef.Field),
			parentName: parentName,
			isAbstract: isAbstract,
		}
	}

	return s[name]
}

// Entity represents schema of the data returned by Microsoft
// fields are inherited from parents so there is a tree hierarchy that can be traversed.
type Entity struct {
	Name       string
	properties map[string]metadatadef.Field
	parentName string
	parent     *Entity
	isAbstract bool
}

func (e *Entity) AddProperty(property metadatadef.Field) {
	e.properties[property.Name] = property
}

// GetRawParentName parents that are defined under schema are prefixed with its alias
// this strips the prefix.
func (e *Entity) GetRawParentName(schemaAlias string) string {
	// strip prefix if we are using local schema
	name, _ := strings.CutPrefix(e.parentName, schemaAlias+".")

	return name
}

// MatchParentsWithChildren this will link a parent entity for everybody
// only top root entities will have parent null.
func (s EntitySet) MatchParentsWithChildren(schemaAlias string) error {
	for _, entity := range s {
		name := entity.GetRawParentName(schemaAlias)
		// if it has parent then match
		if len(name) != 0 {
			parent, ok := s[name]
			if !ok {
				return ErrInvalidParentRef
			}

			entity.parent = parent
		}
	}

	return nil
}

// FilterAbstract returns only concrete entities.
// Abstract schemas are only useful to get inherited properties.
func (s EntitySet) FilterAbstract() EntitySet {
	filtered := make(EntitySet)

	for k, v := range s {
		if !v.isAbstract {
			filtered[k] = v
		}
	}

	return filtered
}

// GetAllProperties recursive function that includes inherited fields from parents.
func (e *Entity) GetAllProperties() map[string]metadatadef.Field {
	if e.parent == nil {
		// this is root
		return e.properties
	}

	parentProperties := e.parent.GetAllProperties()
	combined := make(map[string]metadatadef.Field)

	maps.Copy(combined, e.properties)
	maps.Copy(combined, parentProperties)

	return combined
}

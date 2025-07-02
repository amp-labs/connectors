package snapchatads

const organizationIdKey = "organizationId"

// AuthMetadataVars is a complete list of authentication metadata associated with connector.
// This model serves as a documentation of map[string]string contents.
type AuthMetadataVars struct {
	OrganizationId string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		OrganizationId: dictionary[organizationIdKey],
	}
}

// AsMap converts model back to the map.
func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		organizationIdKey: v.OrganizationId,
	}
}

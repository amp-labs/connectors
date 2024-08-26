package atlassian

const cloudIdKey = "cloudId"

// AuthMetadataVars is a complete list of authentication metadata associated with connector.
// This model serves as a documentation of map[string]string contents.
type AuthMetadataVars struct {
	CloudId string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		CloudId: dictionary[cloudIdKey],
	}
}

// AsMap converts model back to the map.
func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		cloudIdKey: v.CloudId,
	}
}

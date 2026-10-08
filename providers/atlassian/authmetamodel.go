package atlassian

const cloudIdKey = "cloudId"

// AuthMetadataVars is a complete list of authentication metadata associated with connector.
// This model serves as a documentation of map[string]string contents.
type AuthMetadataVars struct {
	CloudId string
}

// AsMap converts model back to the map.
func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		cloudIdKey: v.CloudId,
	}
}

package xero

type AuthMetadataVars struct {
	TenantId string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		TenantId: dictionary["tenantId"],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"tenantId": v.TenantId,
	}
}

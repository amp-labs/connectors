package gusto

type AuthMetadataVars struct {
	CompanyId string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		CompanyId: dictionary[metadataKeyCompanyID],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		metadataKeyCompanyID: v.CompanyId,
	}
}

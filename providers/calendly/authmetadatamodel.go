package calendly

type AuthMetadataVars struct {
	UserURI         string
	OrganizationURI string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(data map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		UserURI:         data["userURI"],
		OrganizationURI: data["organizationURI"],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"userURI":         v.UserURI,
		"organizationURI": v.OrganizationURI,
	}
}

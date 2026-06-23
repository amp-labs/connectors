package zoho

type AuthMetadataVars struct {
	// MailAccountID is the Zoho Mail account id (type ZOHO_ACCOUNT) used in
	// account-scoped Zoho Mail API paths.
	MailAccountID string
}

// NewAuthMetadataVars parses a catalog-variable dictionary into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		MailAccountID: dictionary["zohoMailAccountId"],
	}
}

// AsMap serializes the model into a catalog-variable dictionary.
func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"zohoMailAccountId": v.MailAccountID,
	}
}

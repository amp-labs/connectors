package zoho

// AuthMetadataVars holds the post-authentication metadata persisted for the
// Zoho connector. Currently only the Zoho Mail module populates it, storing
// the account id resolved from the /api/accounts endpoint.
type AuthMetadataVars struct {
	// MailAccountID is the Zoho Mail account id (type ZOHO_ACCOUNT) used in
	// account-scoped Zoho Mail API paths. It is named explicitly for the Mail
	// module so it is not confused with account ids from other Zoho modules.
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

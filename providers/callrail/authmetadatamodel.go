package callrail

type AuthMetadataVars struct {
	AccountID string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(data map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		AccountID: data["accountId"],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"accountId": v.AccountID,
	}
}

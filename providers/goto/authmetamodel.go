package gotoconn

type AuthMetadataVars struct {
	AccountKey string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		AccountKey: dictionary["accountKey"],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"accountKey": v.AccountKey,
	}
}

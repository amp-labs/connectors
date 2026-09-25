package bill

type AuthMetadataVars struct {
	SessionId string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		SessionId: dictionary["sessionId"],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"sessionId": v.SessionId,
	}
}

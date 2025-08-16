package calendly

const (
	userURIKey         = "userURI"
	organizationURIKey = "organizationURI"
)

// AuthMetadataVars represents the metadata variables for Calendly authentication.
// This model serves as documentation of map[string]string contents.
type AuthMetadataVars struct {
	UserURI         string
	OrganizationURI string
}

// AsMap converts model back to the map.
func (a AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		userURIKey:         a.UserURI,
		organizationURIKey: a.OrganizationURI,
	}
} 
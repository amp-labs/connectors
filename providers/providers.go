package providers

// Provider is the name of a provider.
type Provider string

// List all providers here.
const (
	Salesforce Provider = "salesforce"
	Hubspot    Provider = "hubspot"
	LinkedIn   Provider = "linkedin"
)

// String returns the string representation of the provider.
func (p Provider) String() string {
	return string(p)
}

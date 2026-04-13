package providers

// IsSalesforce returns true if the given provider uses the Salesforce APIs.
// This covers the standard OAuth2 Salesforce provider and the headless
// JWT Bearer variant (salesforceJWT). Use this for feature gates that apply
// to any Salesforce-family provider — OAuth-specific gates should still
// check for providers.Salesforce directly.
func IsSalesforce(provider Provider) bool {
	return provider == Salesforce || provider == SalesforceJWT
}

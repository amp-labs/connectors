package providers

// IsSalesforce reports whether the provider is part of the Salesforce family:
// the base provider and its twins, which reuse the same connector
// implementation, APIs and modules, and differ only in authentication scheme
// (salesforceJWT) or in which hosts they address (salesforceCustomDomain).
//
// Prefer this over comparing against Salesforce directly, so that behavior
// gated on "this is Salesforce" reaches every twin. Where a twin is
// deliberately excluded — the OAuth flow specifics, for instance — compare
// against the specific provider so the exclusion is visible.
//
// SalesforceMarketing is not part of this family; it is a separate product with
// its own connector.
func IsSalesforce(provider Provider) bool {
	return provider == Salesforce ||
		provider == SalesforceJWT ||
		provider == SalesforceCustomDomain
}

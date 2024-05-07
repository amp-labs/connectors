package main

import "golang.org/x/oauth2"

// PKCE requires sending challenge during AuthCode and providing verifier during Exchange.
type PKCE struct {
	AuthCode oauth2.AuthCodeOption
	Exchange oauth2.AuthCodeOption
}

func NewPKCE(verifier string) *PKCE {
	return &PKCE{
		AuthCode: oauth2.S256ChallengeOption(verifier),
		Exchange: oauth2.VerifierOption(verifier),
	}
}

// GetAuthOptions provide options for oauth2.AuthCodeURL method.
func (a *OAuthApp) GetAuthOptions() []oauth2.AuthCodeOption {
	if a.PKCE != nil {
		return []oauth2.AuthCodeOption{a.PKCE.AuthCode}
	}

	return nil
}

// GetExchangeOptions provide options for oauth2.Exchange method.
func (a *OAuthApp) GetExchangeOptions() []oauth2.AuthCodeOption {
	if a.PKCE != nil {
		return []oauth2.AuthCodeOption{a.PKCE.Exchange}
	}

	return nil
}

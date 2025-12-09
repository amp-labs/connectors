package proxyserv

import "context"

// CreateProxyOAuth2Password creates Oauth2 Proxy with password grant.
//
// De facto, password grant acts as auth code grant type.
func (f Factory) CreateProxyOAuth2Password(ctx context.Context) *Proxy {
	return f.CreateProxyOAuth2AuthCode(ctx)
}

package paramsbuilder

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

// ParamAssurance checks that param data is valid
// Every param instance must implement it.
type ParamAssurance interface {
	ValidateParams() error
}

var (
	ErrMissingClient    = errors.New("http client not set")
	ErrMissingWorkspace = errors.New("missing workspace name")
)

// Client params sets up authenticated proxy HTTP client
// This can be reused among other param builders by composition.
type Client struct {
	Caller *common.HTTPClient
}

func (p *Client) ValidateParams() error {
	if p.Caller == nil {
		return ErrMissingClient
	}

	return nil
}

func (p *Client) WithClient(
	ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) {
	options := []common.OAuthOption{
		common.WithOAuthClient(client),
		common.WithOAuthConfig(config),
		common.WithOAuthToken(token),
	}

	oauthClient, err := common.NewOAuthHTTPClient(ctx, append(options, opts...)...)
	if err != nil {
		panic(err) // caught in NewConnector
	}

	p.WithAuthenticatedClient(oauthClient)
}

func (p *Client) WithAuthenticatedClient(client common.AuthenticatedHTTPClient) {
	p.Caller = &common.HTTPClient{
		Client:       client,
		ErrorHandler: common.InterpretError,
	}
}

// Workspace params sets up varying workspace name.
type Workspace struct {
	Name string
}

func (p *Workspace) ValidateParams() error {
	if len(p.Name) == 0 {
		return ErrMissingWorkspace
	}

	return nil
}

func (p *Workspace) WithWorkspace(workspaceRef string) {
	p.Name = workspaceRef
}

func (p *Workspace) Substitution() map[string]string {
	return map[string]string{"workspace": p.Name}
}

// Module params adds suffix to URL controlling API versions.
// This is relevant where there are several APIs for different product areas or sub-products, and the APIs
// are versioned differently or have different ways of constructing URLs from object names.
type Module struct {
	Suffix string
}

func (p *Module) ValidateParams() error {
	// url suffix may be omitted
	return nil
}

func (p *Module) WithModule(module APIModule) {
	p.Suffix = module.String()
}

type APIModule struct {
	Label   string // e.g. "crm"
	Version string // e.g. "v3"
}

func (a APIModule) String() string {
	if len(a.Label) == 0 {
		return a.Version
	}

	return fmt.Sprintf("%s/%s", a.Label, a.Version)
}

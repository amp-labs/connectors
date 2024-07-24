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
	ErrMissingClient     = errors.New("http client not set")
	ErrMissingWorkspace  = errors.New("missing workspace name")
	ErrNoSupportedModule = errors.New("no supported module was chosen")
)

// Create will apply options to construct a ready to go set of parameters.
// This is a generalized constructor of parameters used to initialize any connector.
// To qualify as a parameter one must have data validation laid out.
func Create[P ParamAssurance](params P, opts []func(params *P)) (*P, error) {
	for _, opt := range opts {
		opt(&params)
	}

	return &params, params.ValidateParams()
}

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

func (p *Workspace) GetSubstitutionPlan() SubstitutionPlan {
	return SubstitutionPlan{
		From: variableWorkspace,
		To:   p.Name,
	}
}

// Module params adds suffix to URL controlling API versions.
// This is relevant where there are several APIs for different product areas or sub-products, and the APIs
// are versioned differently or have different ways of constructing URLs from object names.
type Module struct {
	Suffix    string
	supported []APIModule
	fallback  *APIModule
}

func (p *Module) ValidateParams() error {
	// making sure the provided module is supported.
	// If the provided module is not supported, use fallback.
	if !p.isSupported() {
		if p.fallback == nil {
			// not supported and user didn't provide a fallback
			return ErrNoSupportedModule
		}

		// replace with fallback module
		p.Suffix = p.fallback.String()

		// even fallback is not supported
		if !p.isSupported() {
			return ErrNoSupportedModule
		}
	}

	return nil
}

func (p *Module) WithModule(module APIModule, supported []APIModule, defaultModule *APIModule) {
	p.Suffix = module.String()
	p.supported = supported
	p.fallback = defaultModule
}

func (p *Module) isSupported() bool {
	for _, mod := range p.supported {
		if p.Suffix == mod.String() {
			return true
		}
	}

	return false
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

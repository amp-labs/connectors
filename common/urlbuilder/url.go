package urlbuilder

import (
	"errors"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

var ErrInvalidURL = errors.New("URL format is incorrect")

// URL focuses on query params and treats the rest as immutable string.
// You can use any URL string to construct this object.
// Its primary goal is to "Expose Query Manipulation".
// Under the hood it uses url.URL library.
type URL struct {
	delegate           *url.URL
	queryParams        url.Values
	encodingExceptions map[string]string
}

// New URL will be constructed given valid full url which may have query params.
func New(base string, path ...string) (*URL, error) {
	delegate, err := url.Parse(cleanTrailingSlashes(base))
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}

	values, err := url.ParseQuery(delegate.RawQuery)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}

	u := &URL{
		delegate:           delegate,
		queryParams:        values,
		encodingExceptions: make(map[string]string),
	}
	u.AddPath(path...)

	return u, nil
}

// FromRawURL converts a core Go `url.URL` into `urlbuilder.URL`,
// providing better control over query parameters and encoding.
func FromRawURL(rawURL *url.URL) (*URL, error) {
	values, err := url.ParseQuery(rawURL.RawQuery)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}

	return &URL{
		delegate:           rawURL,
		queryParams:        values,
		encodingExceptions: nil,
	}, nil
}

func (u *URL) WithQueryParamList(name string, values []string) {
	u.queryParams[name] = values
}

func (u *URL) WithQueryParam(name, value string) {
	u.queryParams[name] = []string{value}
}

func (u *URL) GetFirstQueryParam(name string) (string, bool) {
	value, ok := u.queryParams[name]
	if !ok || len(value) == 0 {
		return "", false
	}

	return value[0], true
}

func (u *URL) RemoveQueryParam(name string) {
	delete(u.queryParams, name)
}

func (u *URL) AddEncodingExceptions(exceptions map[string]string) {
	for key, value := range exceptions {
		u.encodingExceptions[key] = value
	}
}

// ToURL relies on String method.
func (u *URL) ToURL() (*url.URL, error) {
	// Current URL wrapper will be realised as equivalent to url.URL type.
	// It must be done via String() which handles query params.
	result, err := url.Parse(u.String())
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}

	return result, nil
}

func (u *URL) Path() string {
	return u.delegate.Path
}

func (u *URL) String() string {
	// Everything stays the same
	// The only thing that we alter in the delegate's query params
	u.delegate.RawQuery = u.queryValuesToString()

	return u.delegate.String()
}

// URL may have special encoding rules.
// Those can be set via AddEncodingExceptions.
func (u *URL) queryValuesToString() string {
	result := u.queryParams.Encode()
	if len(result) == 0 {
		return ""
	}

	// We are not fully happy with strict encoding provided by url library
	// some special symbols are allowed
	for before, after := range u.encodingExceptions {
		result = strings.ReplaceAll(result, before, after)
	}

	return result
}

func (u *URL) AddPath(paths ...string) *URL {
	// replace delegate with a new URL
	if len(paths) == 0 {
		// nothing to be done here
		return u
	}

	uriParts := make([]string, len(paths))

	for i, p := range paths {
		if i == len(paths)-1 {
			// last index
			p = cleanTrailingSlashes(p)
		}

		uriParts[i] = p
	}

	u.delegate = u.delegate.JoinPath(uriParts...)

	return u
}

func cleanTrailingSlashes(link string) string {
	found := true
	for found {
		link, found = strings.CutSuffix(link, "/")
	}

	return link
}

func (u *URL) HasQueryParam(name string) bool {
	return u.queryParams.Has(name)
}

// Equals compares URL equality ignoring order, encoding.
func (u *URL) Equals(other *URL) bool { // nolint:cyclop
	if strings.ToLower(u.delegate.Host) != strings.ToLower(other.delegate.Host) || // nolint:staticcheck
		u.delegate.Path != other.delegate.Path ||
		u.delegate.RawPath != other.delegate.RawPath ||
		u.delegate.Scheme != other.delegate.Scheme ||
		u.delegate.Fragment != other.delegate.Fragment ||
		u.delegate.RawFragment != other.delegate.RawFragment {
		return false
	}

	// Compare query parameters. The order doesn't matter
	if len(u.queryParams) != len(other.queryParams) {
		return false
	}

	for name, params := range u.queryParams {
		otherParams, exists := other.queryParams[name]
		if !exists || !datautils.EqualUnordered(params, otherParams) {
			return false
		}
	}

	return true
}

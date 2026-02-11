package urlbuilder

import (
	"errors"
	"maps"
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
	delegate *url.URL
	// queryParams has query parameters that should be URL-encoded.
	queryParams url.Values
	// unencodedQueryParams has query parameters that should not be URL-encoded.
	unencodedQueryParams url.Values
	// encodingExceptions are applied to queryParams as string substitutions,
	// after they are encoded.
	// The keys are the strings to be replaced, and the values are the replacements.
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
		delegate:             delegate,
		queryParams:          values,
		unencodedQueryParams: url.Values{},
		encodingExceptions:   make(map[string]string),
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
		delegate:             rawURL,
		queryParams:          values,
		unencodedQueryParams: url.Values{},
		encodingExceptions:   make(map[string]string),
	}, nil
}

func (u *URL) WithQueryParamList(name string, values []string) {
	u.queryParams[name] = values
}

func (u *URL) WithQueryParam(name, value string) {
	u.queryParams[name] = []string{value}
}

// WithUnencodedQueryParam adds a single unencoded query param.
func (u *URL) WithUnencodedQueryParam(name, value string) {
	u.unencodedQueryParams[name] = []string{value}
}

// WithUnencodedQueryParamList adds multiple unencoded query params.
func (u *URL) WithUnencodedQueryParamList(name string, values []string) {
	u.unencodedQueryParams[name] = values
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
	maps.Copy(u.encodingExceptions, exceptions)
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

func (u *URL) Origin() string {
	if u.delegate.Scheme == "" || u.delegate.Host == "" {
		return ""
	}

	return u.delegate.Scheme + "://" + u.delegate.Host
}

func (u *URL) String() string {
	// Everything stays the same
	// The only thing that we alter in the delegate's query params
	u.delegate.RawQuery = u.queryValuesToString()
	output := u.delegate.String()

	return output
}

// URL may have special encoding rules.
// Those can be set via AddEncodingExceptions.
func (u *URL) queryValuesToString() string { // nolint:funcorder
	// Encode the query params
	result := u.queryParams.Encode()
	if len(result) == 0 {
		return ""
	}

	// Apply encoding exceptions
	for before, after := range u.encodingExceptions {
		result = strings.ReplaceAll(result, before, after)
	}

	// Append unencoded params
	if len(u.unencodedQueryParams) > 0 {
		var unencodedParts []string

		for k := range u.unencodedQueryParams {
			vs := u.unencodedQueryParams[k]

			for _, val := range vs {
				unencodedParts = append(unencodedParts, k+"="+val) // no encoding
			}
		}

		if result != "" {
			result += "&" + strings.Join(unencodedParts, "&")
		} else {
			result = strings.Join(unencodedParts, "&")
		}
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

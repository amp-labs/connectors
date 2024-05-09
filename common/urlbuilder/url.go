package urlbuilder

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidURL = errors.New("URL format is incorrect")

// URL focuses on query params and treats the rest as immutable string.
// You can use any URL string to construct this object.
// Its primary goal is to "Expose Query Manipulation".
// Under the hood it uses url.URL library.
type URL struct {
	baseURL            string
	queryParams        url.Values
	fragment           string
	encodingExceptions map[string]string
}

// New URL will be constructed given valid full url which may have query params.
func New(base string) (*URL, error) {
	delegateURL, err := url.Parse(base)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}

	values, err := url.ParseQuery(delegateURL.RawQuery)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}

	result := &URL{
		queryParams: values,
		fragment:    delegateURL.Fragment,
	}

	// Given url.URL structure of:
	// scheme://[userinfo@]host/path[?query][#fragment]
	// omit query and fragment to acquire base of URL
	delegateURL.RawQuery = ""
	delegateURL.RawFragment = ""
	delegateURL.Fragment = ""
	result.baseURL = delegateURL.String()

	return result, nil
}

func (u *URL) WithQueryParamList(name string, values []string) {
	u.queryParams[name] = values
}

func (u *URL) WithQueryParam(name, value string) {
	u.queryParams[name] = []string{value}
}

func (u *URL) RemoveQueryParam(name string) {
	delete(u.queryParams, name)
}

func (u *URL) AddEncodingExceptions(exceptions map[string]string) {
	u.encodingExceptions = exceptions
}

// ToURL relies on String method.
func (u *URL) ToURL() (*url.URL, error) {
	return url.Parse(u.String())
}

func (u *URL) String() string {
	return u.baseURL + u.queryValuesToString() + u.fragmentToString()
}

/*
Return options:

	=> empty string
	=> list of query parameters prefixed with `?`
*/
func (u *URL) queryValuesToString() string {
	// We are not fully happy with strict encoding provided by url library
	// some special symbols are allowed
	result := u.queryParams.Encode()
	if len(result) != 0 {
		for before, after := range u.encodingExceptions {
			result = strings.ReplaceAll(result, before, after)
		}

		return "?" + result
	}

	return ""
}

func (u *URL) fragmentToString() string {
	if len(u.fragment) == 0 {
		return ""
	}

	return "#" + u.fragment
}

func (u *URL) AddPath(paths ...string) *URL {
	delegateURL, _ := url.Parse(u.baseURL)
	delegateURL = delegateURL.JoinPath(paths...)
	u.baseURL = delegateURL.String()

	return u
}

package atlassian

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/golang-jwt/jwt/v5"
)

var jwtExpiryPeriod = 180 * time.Second // nolint:gochecknoglobals
const QuerySeparator = "&"              // nolint:gochecknoglobals

// Documentation:
// https://developer.atlassian.com/cloud/bitbucket/query-string-hash

// JwtTokenGenerator generates the claims on a per-request basis for Atlassian Connect. The JWT needs
// a query request hash, an issued time and an expiration time. The implementation has been adapted
// from https://bitbucket.org/atlassian/atlassian-jwt-js.git.
func JwtTokenGenerator(payload map[string]any, secret string) common.DynamicHeadersGenerator {
	return func(req http.Request) ([]common.Header, error) {
		claims := make(map[string]any)
		queryParams := make(map[string][]string)

		claims["iat"] = time.Now().Unix()
		claims["exp"] = time.Now().Add(jwtExpiryPeriod).Unix()

		for k, v := range req.URL.Query() {
			queryParams[k] = v
		}

		claims["qsh"] = createQueryStringHash(
			req.Method,
			req.URL.Path,
			queryParams,
		)

		// Add any input payload too
		for k, v := range payload {
			claims[k] = v
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			return nil, err
		}

		return []common.Header{
			{
				Key:   "Authorization",
				Value: "JWT " + tokenString,
			},
		}, nil
	}
}

// createQueryStringHash computes the SHA256 hash of the canonical request string.
func createQueryStringHash(
	method,
	path string,
	query map[string][]string,
) string {
	canonicalRequest := createCanonicalRequest(method, path, query)
	hash := sha256.Sum256([]byte(canonicalRequest))

	return hex.EncodeToString(hash[:])
}

// createCanonicalRequest generates the canonical request string.
func createCanonicalRequest(
	method,
	path string,
	query map[string][]string,
) string {
	return canonicalizeMethod(method) +
		QuerySeparator +
		canonicalizeURI(path) +
		QuerySeparator +
		canonicalizeQueryString(query)
}

// canonicalizeMethod converts the HTTP method to uppercase.
func canonicalizeMethod(method string) string {
	return strings.ToUpper(method)
}

func canonicalizeURI(path string) string {
	// Early exit.
	if path == "" {
		return "/"
	}

	// Replace '&' with '%26'.
	path = strings.ReplaceAll(path, "&", "%26")

	// Ensure the path starts with '/'.
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Remove trailing '/' if path length > 1.
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	return path
}

// canonicalizeQueryString generates the canonical query string.
func canonicalizeQueryString(query map[string][]string) string {
	// Step 1: Filter out the 'jwt' parameter.
	filteredQuery := make(map[string][]string)

	for k, v := range query {
		if k != "jwt" {
			filteredQuery[k] = v
		}
	}

	// Step 2: Encode parameter names and values.
	paramStrings := make([]string, 0, len(filteredQuery))

	for paramName, paramValues := range filteredQuery {
		// URL-encode parameter name.
		encodedName := encodeRFC3986(paramName)

		// URL-encode parameter values.
		var encodedValues []string
		for _, val := range paramValues {
			encodedValues = append(encodedValues, encodeRFC3986(val))
		}

		// Step 3: Sort parameter values.
		sort.Strings(encodedValues)

		// Step 4: Join values with ','.
		concatenatedValues := strings.Join(encodedValues, ",")

		// Step 5: Form 'name=value' string.
		paramString := encodedName + "=" + concatenatedValues

		// Collect the parameter string.
		paramStrings = append(paramStrings, paramString)
	}

	// Step 6: Sort parameter strings by encoded parameter names.
	sort.Strings(paramStrings)

	// Step 7: Concatenate parameter strings with '&'.
	canonicalQueryString := strings.Join(paramStrings, "&")

	return canonicalQueryString
}

func encodeRFC3986(str string) string {
	// Use url.QueryEscape to encode special characters.
	encoded := url.QueryEscape(str)

	// Replace '+' with '%20' to encode spaces correctly.
	encoded = strings.ReplaceAll(encoded, "+", "%20")

	return encoded
}

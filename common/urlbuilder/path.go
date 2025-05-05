package urlbuilder

import "strings"

// joinURL appends URI segments to the base URL, preserving slashes where necessary.
//
// This function intentionally avoids using url.URL to prevent any automatic normalization
// or transformations, which are undesired when dealing with relative URI paths
// returned by provider APIs. These paths must be treated as opaque segments.
func joinURL(base string, segments ...string) string {
	// Ensure base doesn't end with a slash.
	joined := strings.TrimRight(base, "/")

	for i, segment := range segments {
		isLast := i == len(segments)-1

		// If last segment ends with slash, preserve it.
		suffixSlash := strings.HasSuffix(segment, "/") && isLast

		// Clean uri segment.
		segment = strings.Trim(segment, "/")

		if len(segment) != 0 {
			joined += "/" + segment
		}

		// Restore trailing slash if it was present on last segment.
		if suffixSlash {
			joined += "/"
		}
	}

	return joined
}

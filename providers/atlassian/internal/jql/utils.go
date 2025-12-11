package jql

import (
	"fmt"
	"time"
)

/*

	Querying in JQL is expressed in *relative time* from `now`.

	`now` is treated as the origin (0 point on the timeline).
	All bounds — such as `since` or `until` — are measured in minutes before `now`.

	Below are examples that show how an absolute time range is converted into a JQL query string.

	Dashes represent minutes on a timeline:

	(1) since----until---NOW
		* `since` is 7 minutes before now	=> -7m
		* `until` is 7 minutes before now	=> -3m
		Final query: updated > "-7m" AND updated < "-3m"

	(2) since----NOW---until
    	- `since` is 4 minutes before now	=> -4m
    	- `until` is 3 minutes after now	=> +3m (defaults to 0)
		Final query: updated > "-4m" AND updated < "0"

	(3) NOW----since---until
    	- `since` is 4 minutes after now	=> +4m (defaults to 0)
    	- `until` is 7 minutes after now	=> +7m (defaults to 0)
		Final query: updated > "0" AND updated < "0"

*/

// RelativeMinutesFromNow returns the number of minutes the given timestamp is in the past,
// relative to the current moment (`now`), as a negative integer string. Ex: "-5m", 5 minutes before now.
// If the timestamp is in the future or at the present moment, it returns "0".
// If timestamp is zero, it returns an empty string, meaning no query should be applied.
func relativeMinutesFromNow(timestamp time.Time) string {
	if timestamp.IsZero() {
		return ""
	}

	diff := time.Since(timestamp)

	// Any timestamp in the future is treated as "now" (0),
	// because querying for records "since the future" is illogical.
	// Until suffers from the same problem. Records are always created before right now.
	// The latest until could be is now.
	minutes := int64(diff.Minutes())
	if minutes <= 0 {
		return "0"
	}

	return fmt.Sprintf("-%v", minutes)
}

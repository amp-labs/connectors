package naming

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var ErrInvalidJSONFormat = errors.New("invalid JSON type format")

// Text acts universally as either JSON string or JSON int.
type Text string

func (t *Text) UnmarshalJSON(data []byte) error {
	// Try as a string.
	var s string

	if err := json.Unmarshal(data, &s); err == nil {
		*t = Text(s)

		return nil
	}

	// Try as a number.
	var number float64

	if err := json.Unmarshal(data, &number); err == nil {
		// Check if the number is whole or has decimals.
		if number == float64(int64(number)) {
			// no decimal
			*t = Text(strconv.FormatInt(int64(number), 10))
		} else {
			// preserve decimal
			*t = Text(strconv.FormatFloat(number, 'f', -1, 64))
		}

		return nil
	}

	return fmt.Errorf("%w: %s", ErrInvalidJSONFormat, string(data))
}

func (t *Text) String() string {
	return string(*t)
}

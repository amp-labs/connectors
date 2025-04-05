package monday

import (
	"fmt"
)

type ResponseMessageError struct {
	Message string `json:"message"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Message) != 0 {
		return fmt.Errorf("%w: %s", base, r.Message)
	}

	return base
}

type ResponseBasicError struct {
	Error string `json:"error"`
}

func (r ResponseBasicError) CombineErr(base error) error {
	if len(r.Error) != 0 {
		return fmt.Errorf("%w: %s", base, r.Error)
	}

	return base
}

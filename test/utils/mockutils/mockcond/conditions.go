package mockcond

import "net/http"

// Condition computes whether http.Request meets some rule.
type Condition interface {
	EvaluateCondition(w http.ResponseWriter, r *http.Request) bool
}

// Check is the most basic Condition. It is a function definition that allows custom implementation.
// There are some out of the box functions in this package that have this signature.
type Check func(w http.ResponseWriter, r *http.Request) bool

// Or is a composite Condition which evaluates to true if at least one condition is met.
// Empty list returns false.
type Or []Condition

// And is a composite Condition which evaluates to true if all conditions are met.
// Empty list returns false.
type And []Condition

func (c Check) EvaluateCondition(w http.ResponseWriter, r *http.Request) bool {
	return c(w, r)
}

func (o Or) EvaluateCondition(w http.ResponseWriter, r *http.Request) bool {
	if len(o) == 0 {
		return false
	}

	for _, condition := range o {
		if condition == nil {
			// Nil conditions are not allowed.
			return false
		}

		if condition.EvaluateCondition(w, r) {
			return true
		}
	}

	return false
}
func (a And) EvaluateCondition(w http.ResponseWriter, r *http.Request) bool {
	if len(a) == 0 {
		return false
	}

	for _, condition := range a {
		if condition == nil {
			// Nil conditions are not allowed.
			return false
		}

		if !condition.EvaluateCondition(w, r) {
			return false
		}
	}

	return true
}

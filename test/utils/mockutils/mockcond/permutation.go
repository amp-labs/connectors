package mockcond

// Permute generates an Or condition containing one branch for every
// permutation of the provided items. For each permutation, the
// conditionMaker function is invoked with a slice representing that
// specific ordering, and the result is included in the returned Or.
//
// This is useful when a mocked API must match an exact string whose
// internal ordering may vary. For example, if a connector builds a
// query like:
//
//	"SELECT A,B,C FROM contacts"
//
// but the order of fields ("A,B,C") is not guaranteed, Permute allows
// tests to enumerate all possible valid variants without writing them
// manually.
//
// Example usage:
//
//	// Custom constructor for the test suite.
//	queryParam := func(template string) func(fields []string) mockcond.Condition {
//	    return func(fields []string) mockcond.Condition {
//	        selector := strings.Join(fields, ",")
//	        return mockcond.QueryParam("q", fmt.Sprintf(template, selector))
//	    }
//	}
//
//	cond := mockcond.Permute(
//	    queryParam("SELECT %v FROM contacts"),
//	    "AssistantName", "Department",
//	)
//
// The resulting Or condition will match:
//
//	SELECT AssistantName,Department FROM contacts
//	SELECT Department,AssistantName FROM contacts
//
// Parameters:
//
//	conditionMaker — a function that receives one permutation of items
//	                 and returns a Condition to test that permutation
//
//	items — the elements to permute; providing N items produces N!
//	        conditions, so use carefully with large sets
//
// Returns:
//
//	An Or containing one Condition per permutation. If items is empty,
//	the returned Or is also empty and will evaluate to false.
func Permute[T comparable](
	conditionMaker func([]T) Condition,
	items ...T,
) Or {
	perms := generatePermutations(items)

	oneOf := Or{}
	for _, permutation := range perms {
		oneOf = append(oneOf, conditionMaker(permutation))
	}

	return oneOf
}

// generatePermutations returns all permutations of the values inside the set.
// The returned slice contains permutations in an arbitrary order, but includes
// every possible ordering exactly once.
func generatePermutations[T comparable](items []T) [][]T {
	// Convert the set to a slice, because permutation algorithms need indexable collections.
	size := len(items)
	if size == 0 {
		return [][]T{}
	}

	// Make a working copy to mutate during permutation.
	array := make([]T, size)
	copy(array, items)

	var result [][]T

	// Heap's algorithm
	var generate func(int)
	generate = func(k int) {
		if k == 1 {
			perm := make([]T, size)
			copy(perm, array)
			result = append(result, perm)
			return
		}

		for i := 0; i < k; i++ {
			generate(k - 1)

			// Swap depending on parity of k
			if k%2 == 1 {
				array[0], array[k-1] = array[k-1], array[0]
			} else {
				array[i], array[k-1] = array[k-1], array[i]
			}
		}
	}

	generate(size)

	return result
}

package common

// FilterBy returns a new SearchFilter with an additional FieldFilter appended.
//
// The receiver is treated as immutable: FilterBy does not modify the original
// SearchFilter. Instead, it creates a copy of the underlying FieldFilters slice
// before appending the new filter. This avoids slice aliasing and allows safe
// chaining of FilterBy calls without unintended side effects.
//
// Example:
//
//	filter := SearchFilter{}.
//		FilterBy("status", Eq, "active").
//		FilterBy("age", Gt, 18)
func (s SearchFilter) FilterBy(fieldName string, operator FilterOperator, value any) SearchFilter {
	// Copy the slice to avoid aliasing
	newFilters := make([]FieldFilter, len(s.FieldFilters))
	copy(newFilters, s.FieldFilters)

	newFilters = append(newFilters, FieldFilter{
		FieldName: fieldName,
		Operator:  operator,
		Value:     value,
	})

	return SearchFilter{FieldFilters: newFilters}
}

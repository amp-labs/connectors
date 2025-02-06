package datautils

// IndexedLists is a dynamic list of identifiable slices.
// Each slice is associated to the unique comparable object.
type IndexedLists[ID comparable, V any] map[ID][]V

// NamedLists is a collection of lists each having a string name.
type NamedLists[V any] IndexedLists[string, V]

// UniqueLists is a collection of Sets each associated to identifier.
type UniqueLists[ID, V comparable] map[ID]Set[V]

// Add objects into the named list.
// If the list didn't exist it will create it.
func (l IndexedLists[ID, V]) Add(bucket ID, objects ...V) {
	if _, ok := l[bucket]; !ok {
		l[bucket] = make([]V, 0)
	}

	l[bucket] = append(l[bucket], objects...)
}

func (l NamedLists[V]) Add(bucket string, objects ...V) {
	IndexedLists[string, V](l).Add(bucket, objects...)
}

func (l UniqueLists[ID, V]) Add(bucket ID, objects ...V) {
	if _, ok := l[bucket]; !ok {
		l[bucket] = NewSet[V]()
	}

	l[bucket].Add(objects)
}

func (l UniqueLists[ID, V]) GetObjects(bucket ID) []V {
	return l[bucket].List()
}

// GetBuckets returns names of lists, buckets they are in.
// Example:
//
//	veggies: cucumber, tomato;
//	fruits: pineapple;
//
// Will return veggies and fruits.
func (l IndexedLists[ID, V]) GetBuckets() []ID {
	result := make([]ID, len(l))

	index := 0

	for name := range l {
		result[index] = name
		index++
	}

	return result
}

func (l NamedLists[V]) GetBuckets() []string {
	return IndexedLists[string, V](l).GetBuckets()
}

func (l UniqueLists[ID, V]) GetBuckets() []ID {
	result := make([]ID, len(l))

	index := 0

	for name := range l {
		result[index] = name
		index++
	}

	return result
}

func (l IndexedLists[ID, V]) MergeWith(other IndexedLists[ID, V]) {
	for k, v := range other {
		l.Add(k, v...)
	}
}

func (l NamedLists[V]) MergeWith(other NamedLists[V]) {
	IndexedLists[string, V](l).MergeWith(IndexedLists[string, V](other))
}

package utils

import "sort"

// SortableByIndex is an interface for objects that have an Index field.
type SortableByIndex interface {
	GetIndex() int64
}

// SortByIndex sorts a slice of SortableByIndex objects by their Index field.
func SortByIndex[T SortableByIndex](slice []T) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].GetIndex() < slice[j].GetIndex()
	})
}

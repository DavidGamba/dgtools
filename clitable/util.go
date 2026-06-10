package clitable

import "slices"

// Given a base slice, it allows to introduce a given prefix slice with a fixed ordering.
// It will prefix the base slice with the entries from prefix and drop the entries from base.
// If a prefix entry doesn't exist in base, it ignores it.
//
// Useful when you have an ordered slice of map keys but you can to provide a way to access some of the map keys in an ordered way.
func prefixSlice[T comparable](prefix, base []T) []T {
	if len(prefix) <= 0 {
		return base
	}
	validPrefix := []T{}
	for _, e := range prefix {
		index := slices.Index(base, e)
		if index < 0 {
			continue
		}
		validPrefix = append(validPrefix, e)
		base = slices.Delete(base, index, index+1)
	}
	return append(validPrefix, base...)
}

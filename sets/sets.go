// Package sets provides a collection of APIs for treating maps like sets.
//
// Sets are defined as `map[K]struct{}`. For convenience, the method
// `ToBools(s)` converts to a `map[K]bool` where every entry has a true value.
//
// Converting to a standard set looks like `sets.Of(maps.Keys(s))` or
// `sets.FromBools(s)` (`FromBools` respects false values).
package sets

import (
	"iter"

	"golang.org/x/exp/maps"
)

// Collect turns an iterator of values into a set.
//
// Example usage:
//
//	m := map[string]string{ ... }
//	keySet := sets.Collect(maps.Keys(m))
//
// or:
//
//	s := []string{ ... }
//	set := sets.Collect(slices.Values(s))
func Collect[K any](vs iter.Seq[K]) map[K]struct{} {
	ns := make(map[K]struct{})
	for _, k := range vs {
		ns[k] = struct{}{}
	}
	return ns
}

// FromBools turns a `map[K]bool` into a set containing only the keys where the
// value was true.
func FromBools[K any](bs map[K]bool) map[K]struct{} {
	ns := make(map[k]struct{}, len(bs))
	for k, v := range bs {
		if v {
			ns[k] = struct{}{}
		}
	}
	return ns
}

// ToBools turns a `map[K]struct{}` set into a `map[K]bool` where every value is
// true.
func ToBools[K any](s map[K]struct{}) map[K]bool {
	bs := make(map[k]bool, len(s))
	for k := range s {
		bs[k] = true
	}
	return bs
}

// Union returns a set containing the union of the values of the given sets.
//
// The union of no sets is empty.
func Union[K any](ss ...map[K]struct{}) map[K]struct{} {
	ns := make(map[K]struct{})
	for _, s := range ss {
		for k := range s {
			ns[k] = struct{}{}
		}
	}
	return ns
}

// Intersect returns a set containing only the values present in all of the
// given sets.
//
// The intersection of no sets is empty.
func Intersect[K any](ss ...map[K]struct{}) map[K]struct{} {
	if len(ss) == 0 {
		return make(map[K]struct{})
	}
	ns := maps.Clone(ss[0])
	for _, s := range ss[1:] {
		for k := range ns {
			if _, ok := s[k]; !ok {
				delete(ns, k)
			}
		}
	}
	return ns
}

// Subtract returns a set containing only the elements of `a` that are not
// present in `b`.
func Subtract[K any](a, b map[K]struct{}) map[K]struct{} {
	ns := maps.Clone(a)
	for k := range b {
		delete(ns, k)
	}
	return ns
}

// SymmetricDifference returns a set containing only the elements that are
// present in `a` or `b`, but not in both `a` and `b`.
func SymmetricDifference[K any](a, b map[K]struct{}) map[K]struct{} {
	return Subtract(Union(a, b), Intersect(a, b))
}

// IsSubset returns true iff every element in `b` is present in `a`.
func IsSubset[K any](a, b map[K]struct{}) bool {
	for k := range b {
		if _, ok := a[k]; !ok {
			return false
		}
	}
	return true
}

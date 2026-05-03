// Package sets provides a collection of APIs for treating maps like sets.
//
// Sets are defined as `map[K]struct{}`. For convenience, `ToBools(s)` converts
// to a `map[K]bool` where every entry has a true value.
//
// Converting to a set might look like `sets.CollectKeys(m)`,
// `sets.FromBools(mapToBool)`, or `sets.Collect(iterator)`.
package sets

import (
	"iter"

	"maps"
	"slices"
)

// Equal returns true iff `a` and `b` contain the same elements.
func Equal[K comparable](a, b map[K]struct{}) bool { return maps.Equal(a, b) }

// Clone returns a shallow copy of the given set.
//
// Returns nil if the given set was nil.
func Clone[K comparable](s map[K]struct{}) map[K]struct{} { return maps.Clone(s) }

// CollectSlice turns a slice into a set.
//
// Never returns nil, even for empty/nil input.
func CollectSlice[K comparable](s []K) map[K]struct{} {
	return Collect(slices.Values(s))
}

// CollectKeys turns the keys of a map into a set.
//
// Never returns nil, even for empty/nil input.
func CollectKeys[K comparable, V any](m map[K]V) map[K]struct{} {
	return Collect(maps.Keys(m))
}

// Collect turns an iterator of values into a set.
//
// For convenience, consider using `CollectSlice()` or `CollectKeys()` instead.
//
// Never returns nil, even for empty input.
func Collect[K comparable](vs iter.Seq[K]) map[K]struct{} {
	ns := make(map[K]struct{})
	for k := range vs {
		ns[k] = struct{}{}
	}
	return ns
}

// FromBools turns a `map[K]bool` into a set containing only the keys where the
// value was true.
//
// Never returns nil, even for empty/nil input.
func FromBools[K comparable](bs map[K]bool) map[K]struct{} {
	ns := make(map[K]struct{}, len(bs))
	for k, v := range bs {
		if v {
			ns[k] = struct{}{}
		}
	}
	return ns
}

// ToBools turns a `map[K]struct{}` set into a `map[K]bool` where every value is
// true.
//
// Never returns nil, even for empty/nil input.
func ToBools[K comparable](s map[K]struct{}) map[K]bool {
	bs := make(map[K]bool, len(s))
	for k := range s {
		bs[k] = true
	}
	return bs
}

// Union returns a set containing the union of the values of the given sets.
//
// The union of no sets is empty.
// Never returns nil, even for empty/nil input.
func Union[K comparable](ss ...map[K]struct{}) map[K]struct{} {
	ns := make(map[K]struct{})
	for _, s := range ss {
		for k := range s {
			ns[k] = struct{}{}
		}
	}
	return ns
}

// UnionSeq returns a set containing the union of the values of the given
// sequences.
//
// The union of no sequences is empty.
// Never returns nil, even for empty input.
func UnionSeq[K comparable](ss ...iter.Seq[K]) map[K]struct{} {
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
// Never returns nil, even for empty/nil input.
func Intersect[K comparable](ss ...map[K]struct{}) map[K]struct{} {
	if len(ss) == 0 || ss[0] == nil {
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
//
// Never returns nil, even for empty/nil input.
func Subtract[K comparable](a, b map[K]struct{}) map[K]struct{} {
	if a == nil {
		return make(map[K]struct{})
	}
	ns := maps.Clone(a)
	for k := range b {
		delete(ns, k)
	}
	return ns
}

// SymmetricDifference returns a set containing only the elements that are
// present in `a` or `b`, but not in both `a` and `b`.
//
// Never returns nil, even for empty/nil input.
func SymmetricDifference[K comparable](a, b map[K]struct{}) map[K]struct{} {
	return Subtract(Union(a, b), Intersect(a, b))
}

// Overlaps returns true iff `a` and `b` share at least one element.
func Overlaps[K comparable](a, b map[K]struct{}) bool {
	if len(a) > len(b) {
		a, b = b, a // Range over the smaller set for efficiency.
	}
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return false
}

// IsSubset returns true iff every element in `sub` is present in `super`.
func IsSubset[K comparable](super, sub map[K]struct{}) bool {
	for k := range sub {
		if _, ok := super[k]; !ok {
			return false
		}
	}
	return true
}

// InsertAll inserts every element from each of `ss` into `s`.
//
// If `s` is nil, InsertAll allocates and returns a new set.
// Never returns nil.
func InsertAll[K comparable](s map[K]struct{}, ss ...map[K]struct{}) map[K]struct{} {
	if s == nil {
		s = make(map[K]struct{})
	}
	for _, other := range ss {
		for k := range other {
			s[k] = struct{}{}
		}
	}
	return s
}

// Filter removes from `s` every element for which `keep` returns false.
//
// Never returns nil, even for empty/nil input.
func Filter[K comparable](s map[K]struct{}, keep func(K) bool) map[K]struct{} {
	if s == nil {
		return make(map[K]struct{})
	}
	for k := range s {
		if !keep(k) {
			delete(s, k)
		}
	}
	return s
}

// Pop removes and returns an arbitrary element from `s`.
//
// If `s` is empty or nil, Pop returns the zero value of K and false.
func Pop[K comparable](s map[K]struct{}) (k K, ok bool) {
	for k := range s {
		delete(s, k)
		return k, true
	}
	return
}

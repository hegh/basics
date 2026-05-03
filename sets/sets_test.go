package sets

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEqual(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
		1: {},
	}

	if !Equal(a, b) {
		t.Fatal("got false want true")
	}
}

func TestEqualFalse(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
		3: {},
	}

	if Equal(a, b) {
		t.Fatal("got true want false")
	}
}

func TestClone(t *testing.T) {
	s := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Clone(s)
	if diff := cmp.Diff(s, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestCloneDoesNotAliasInput(t *testing.T) {
	s := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Clone(s)
	delete(got, 1)

	if _, ok := s[1]; !ok {
		t.Fatalf("got %#v want original input preserved for alias check", s)
	}
}

func TestCloneNil(t *testing.T) {
	got := Clone[int](nil)
	if got != nil {
		t.Fatalf("got %#v want nil", got)
	}
}

func TestCollect(t *testing.T) {
	got := Collect(slices.Values([]int{1, 2, 2, 3}))
	want := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestCollectEmpty(t *testing.T) {
	got := Collect(slices.Values([]int{}))
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestCollectSlice(t *testing.T) {
	got := CollectSlice([]int{1, 2, 2, 3})
	want := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestCollectSliceNil(t *testing.T) {
	got := CollectSlice[int](nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestCollectKeys(t *testing.T) {
	got := CollectKeys(map[string]int{
		"a": 1,
		"b": 2,
	})
	want := map[string]struct{}{
		"a": {},
		"b": {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestCollectKeysNil(t *testing.T) {
	got := CollectKeys[string, int](nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestFromBools(t *testing.T) {
	got := FromBools(map[string]bool{
		"a": true,
		"b": false,
		"c": true,
	})
	want := map[string]struct{}{
		"a": {},
		"c": {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestFromBoolsNil(t *testing.T) {
	got := FromBools[string](nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestFromBoolsEmpty(t *testing.T) {
	got := FromBools(map[string]bool{})
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestToBools(t *testing.T) {
	got := ToBools(map[string]struct{}{
		"a": {},
		"b": {},
	})
	want := map[string]bool{
		"a": true,
		"b": true,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestToBoolsNil(t *testing.T) {
	got := ToBools[string](nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestToBoolsEmpty(t *testing.T) {
	got := ToBools(map[string]struct{}{})
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestUnion(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
		3: {},
	}
	c := map[int]struct{}{
		4: {},
	}

	got := Union(a, b, c)
	want := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestUnionDoesNotMutateInputs(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
		3: {},
	}

	_ = Union(a, b)

	wantA := map[int]struct{}{
		1: {},
		2: {},
	}
	wantB := map[int]struct{}{
		2: {},
		3: {},
	}
	if diff := cmp.Diff(wantA, a); diff != "" {
		t.Fatalf("got diff (-want +got) for first input after Union():\n%s", diff)
	}
	if diff := cmp.Diff(wantB, b); diff != "" {
		t.Fatalf("got diff (-want +got) for second input after Union():\n%s", diff)
	}
}

func TestUnionSingleSetDoesNotAliasInput(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Union(a)
	delete(got, 1)

	if _, ok := a[1]; !ok {
		t.Fatalf("got %#v want original input preserved for alias check", a)
	}
}

func TestUnionNoSets(t *testing.T) {
	got := Union[int]()
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestUnionNilInputs(t *testing.T) {
	got := Union[int](nil, nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestUnionSeq(t *testing.T) {
	got := UnionSeq(
		slices.Values([]int{1, 2, 2}),
		slices.Values([]int{2, 3}),
		slices.Values([]int{4}),
	)
	want := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestUnionSeqNoSequences(t *testing.T) {
	got := UnionSeq[int]()
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestIntersect(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	b := map[int]struct{}{
		2: {},
		3: {},
		4: {},
	}
	c := map[int]struct{}{
		0: {},
		2: {},
		3: {},
	}

	got := Intersect(a, b, c)
	want := map[int]struct{}{
		2: {},
		3: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestIntersectNoSets(t *testing.T) {
	got := Intersect[int]()
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestIntersectNilInput(t *testing.T) {
	got := Intersect[int](nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestIntersectSingleSet(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Intersect(a)
	if diff := cmp.Diff(a, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestIntersectDoesNotMutateInput(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
	}

	_ = Intersect(a, b)

	wantA := map[int]struct{}{
		1: {},
		2: {},
	}
	if diff := cmp.Diff(wantA, a); diff != "" {
		t.Fatalf("got diff (-want +got) for first input after Intersect():\n%s", diff)
	}
}

func TestIntersectSingleSetDoesNotAliasInput(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Intersect(a)
	delete(got, 1)

	if _, ok := a[1]; !ok {
		t.Fatalf("got %#v want original input preserved for alias check", a)
	}
}

func TestSubtract(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	b := map[int]struct{}{
		2: {},
		4: {},
	}

	got := Subtract(a, b)
	want := map[int]struct{}{
		1: {},
		3: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestSubtractDoesNotMutateInput(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
	}

	_ = Subtract(a, b)

	wantA := map[int]struct{}{
		1: {},
		2: {},
	}
	if diff := cmp.Diff(wantA, a); diff != "" {
		t.Fatalf("got diff (-want +got) for first input after Subtract():\n%s", diff)
	}
}

func TestSubtractDoesNotAliasFirstInput(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Subtract(a, map[int]struct{}{})
	delete(got, 1)

	if _, ok := a[1]; !ok {
		t.Fatalf("got %#v want original input preserved for alias check", a)
	}
}

func TestSubtractNilInputs(t *testing.T) {
	got := Subtract[int](nil, nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestSymmetricDifference(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	b := map[int]struct{}{
		3: {},
		4: {},
		5: {},
	}

	got := SymmetricDifference(a, b)
	want := map[int]struct{}{
		1: {},
		2: {},
		4: {},
		5: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestSymmetricDifferenceIdenticalSets(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		1: {},
		2: {},
	}

	got := SymmetricDifference(a, b)
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestSymmetricDifferenceWithEmptySet(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	got := SymmetricDifference(a, map[int]struct{}{})
	if diff := cmp.Diff(a, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestSymmetricDifferenceDoesNotAliasInput(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	got := SymmetricDifference(a, map[int]struct{}{})
	delete(got, 1)

	if _, ok := a[1]; !ok {
		t.Fatalf("got %#v want original input preserved for alias check", a)
	}
}

func TestSymmetricDifferenceNilInputs(t *testing.T) {
	got := SymmetricDifference[int](nil, nil)
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestIsSubset(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	b := map[int]struct{}{
		1: {},
		3: {},
	}

	if !IsSubset(a, b) {
		t.Fatal("got false want true")
	}
}

func TestIsSubsetFalse(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}
	b := map[int]struct{}{
		2: {},
		3: {},
	}

	if IsSubset(a, b) {
		t.Fatal("got true want false")
	}
}

func TestIsSubsetEmptySet(t *testing.T) {
	a := map[int]struct{}{
		1: {},
		2: {},
	}

	if !IsSubset(a, map[int]struct{}{}) {
		t.Fatal("got false want true for empty subset")
	}
}

func TestInsertAll(t *testing.T) {
	s := map[int]struct{}{
		1: {},
	}

	got := InsertAll(s,
		map[int]struct{}{
			2: {},
		},
		map[int]struct{}{
			2: {},
			3: {},
		},
	)
	want := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want, s); diff != "" {
		t.Fatalf("got diff (-want +got) for input after InsertAll():\n%s", diff)
	}
}

func TestInsertAllNilSet(t *testing.T) {
	got := InsertAll[int](nil, map[int]struct{}{
		1: {},
		2: {},
	})
	want := map[int]struct{}{
		1: {},
		2: {},
	}
	if got == nil {
		t.Fatal("got nil want non-nil map")
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
}

func TestFilter(t *testing.T) {
	s := map[int]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
	}

	got := Filter(s, func(k int) bool {
		return k%2 == 0
	})
	want := map[int]struct{}{
		2: {},
		4: {},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("got diff (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want, s); diff != "" {
		t.Fatalf("got diff (-want +got) for input after Filter():\n%s", diff)
	}
}

func TestFilterNilInput(t *testing.T) {
	got := Filter[int](nil, func(int) bool {
		return true
	})
	if got == nil {
		t.Fatal("got nil want empty non-nil map")
	}
	if len(got) != 0 {
		t.Fatalf("got len %d want 0", len(got))
	}
}

func TestFilterReturnsMutatedInput(t *testing.T) {
	s := map[int]struct{}{
		1: {},
		2: {},
	}

	got := Filter(s, func(int) bool {
		return true
	})
	if got == nil {
		t.Fatal("got nil want non-nil map")
	}
	delete(got, 1)

	if _, ok := s[1]; ok {
		t.Fatalf("got %#v want input mutated through returned map", s)
	}
}

func TestPop(t *testing.T) {
	s := map[int]struct{}{
		1: {},
	}

	got, ok := Pop(s)
	if !ok {
		t.Fatal("got false want true")
	}
	if got != 1 {
		t.Fatalf("got %d want 1", got)
	}
	if len(s) != 0 {
		t.Fatalf("got len %d want 0 after Pop()", len(s))
	}
}

func TestPopEmpty(t *testing.T) {
	got, ok := Pop(map[int]struct{}{})
	if ok {
		t.Fatal("got true want false")
	}
	if got != 0 {
		t.Fatalf("got %d want 0", got)
	}
}

func TestPopNil(t *testing.T) {
	got, ok := Pop[int](nil)
	if ok {
		t.Fatal("got true want false")
	}
	if got != 0 {
		t.Fatalf("got %d want 0", got)
	}
}

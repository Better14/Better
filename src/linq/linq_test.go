// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq_test

import (
	"linq"
	"slices"
	"testing"
)

func sliceCount[T any](s []T) (int, bool) {
	return s.TryGetNonEnumeratedCount()
}

func TestChainWhereSelectToList(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	out := linq.From(nums).Where(func(n int) bool { return n < 5 }).Select(func(n int) int { return n + 1 }).ToList()
	want := []int{2, 3, 4, 5}
	if !slices.Equal(out, want) {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestChainFirst(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := linq.From(nums).Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestLazyPackageAPI(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := linq.From(nums).Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestLazyReceiverSelect(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := linq.From(nums).Where(func(n int) bool { return n%2 == 0 }).Select(func(n int) int { return n * 2 }).First()
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestSumChain(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sum := linq.From(nums).Where(func(n int) bool { return n%2 == 0 }).Sum()
	if sum != 6 {
		t.Fatalf("sum: %v", sum)
	}
}

func TestRangeRepeatEmpty(t *testing.T) {
	got := linq.Range(2, 3).ToList()
	want := []int{2, 3, 4}
	if !slices.Equal(got, want) {
		t.Fatalf("Range: got %v", got)
	}
	if linq.Repeat(7, 2).First() != 7 {
		t.Fatal("Repeat")
	}
	if linq.Empty[int]().Any(func(int) bool { return true }) {
		t.Fatal("Empty")
	}
}

func TestSetOperations(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := []int{3, 4, 5, 6}
	union := linq.Union(linq.From(a), linq.From(b)).ToList()
	if len(union) != 6 {
		t.Fatalf("Union: %v", union)
	}
	except := linq.Except(linq.From(a), linq.From(b)).ToList()
	if !slices.Equal(except, []int{1, 2}) {
		t.Fatalf("Except: %v", except)
	}
	inter := linq.Intersect(linq.From(a), linq.From(b)).ToList()
	if !slices.Equal(inter, []int{3, 4}) {
		t.Fatalf("Intersect: %v", inter)
	}
}

func TestOrderThenBy(t *testing.T) {
	type person struct {
		name string
		age  int
	}
	people := []person{{"bob", 30}, {"alice", 25}, {"alice", 20}}
	out := linq.From(people).OrderBy(func(p person) string { return p.name }).ThenBy(func(p person) int { return p.age }).ToList()
	if out[0].age != 20 || out[2].name != "bob" {
		t.Fatalf("Order/ThenBy: %v", out)
	}
}

func TestJoin(t *testing.T) {
	outer := []linq.KeyValue[string, int]{{"a", 1}, {"b", 2}}
	inner := []linq.KeyValue[string, string]{{"a", "x"}, {"b", "y"}}
	joined := linq.From(outer).Join(
		linq.From(inner),
		func(kv linq.KeyValue[string, int]) string { return kv.Key },
		func(kv linq.KeyValue[string, string]) string { return kv.Key },
		func(o linq.KeyValue[string, int], i linq.KeyValue[string, string]) string {
			return o.Key + i.Value
		},
	).ToList()
	if !slices.Equal(joined, []string{"ax", "by"}) {
		t.Fatalf("Join: %v", joined)
	}
}

func TestMaterializers(t *testing.T) {
	nums := []int{1, 2, 2, 3}
	seq := linq.From(nums)
	dict := linq.ToDictionary(seq,
		func(n int) int { return n },
		func(n int) string { return "v" },
	)
	if len(dict) != 3 {
		t.Fatalf("ToDictionary: %v", dict)
	}
	lookup := linq.ToLookup(seq,
		func(n int) int { return n % 2 },
		func(n int) int { return n },
	)
	if lookup.Count() != 2 || len(lookup.Get(0)) != 1 {
		t.Fatalf("ToLookup: %v", lookup.Get(0))
	}
	set := linq.ToHashSet(seq)
	if len(set) != 3 {
		t.Fatalf("ToHashSet: %v", set)
	}
}

func TestTryGetNonEnumeratedCount(t *testing.T) {
	nums := []int{1, 2, 3}
	n, ok := sliceCount(nums)
	if !ok || n != 3 {
		t.Fatalf("got (%d, %v)", n, ok)
	}
	_, ok = linq.From([]int{1, 2, 3}).Where(func(int) bool { return true }).TryGetNonEnumeratedCount()
	if ok {
		t.Fatal("expected unknown count after Where")
	}
}

func TestTerminalOps(t *testing.T) {
	nums := []int{3, 1, 4, 1, 5}
	seq := linq.From(nums)
	if seq.Max() != 5 || seq.Min() != 1 {
		t.Fatal("Min/Max")
	}
	if seq.Average() != 2.8 {
		t.Fatalf("Average: %v", seq.Average())
	}
	if seq.Last() != 5 || seq.ElementAt(2) != 4 {
		t.Fatal("Last/ElementAt")
	}
	if linq.From(nums).Where(func(n int) bool { return n == 3 }).Single() != 3 {
		t.Fatal("Single")
	}
}

func TestZipChunkAppendPrepend(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{10, 20, 30}
	zipped := linq.From(a).Zip(linq.From(b), func(x, y int) int { return x + y }).ToList()
	if !slices.Equal(zipped, []int{11, 22, 33}) {
		t.Fatalf("Zip: %v", zipped)
	}
	chunks := linq.From(a).Chunk(2).ToList()
	if len(chunks) != 2 || len(chunks[1]) != 1 {
		t.Fatalf("Chunk: %v", chunks)
	}
	appended := linq.From(a).Append(4).ToList()
	if !slices.Equal(appended, []int{1, 2, 3, 4}) {
		t.Fatalf("Append: %v", appended)
	}
	prepended := linq.From(a).Prepend(0).ToList()
	if !slices.Equal(prepended, []int{0, 1, 2, 3}) {
		t.Fatalf("Prepend: %v", prepended)
	}
}

func TestCastOfType(t *testing.T) {
	src := linq.From([]any{1, "x", 2, 3.0})
	ints := src.OfType[int]().ToList()
	if !slices.Equal(ints, []int{1, 2}) {
		t.Fatalf("OfType: %v", ints)
	}
}

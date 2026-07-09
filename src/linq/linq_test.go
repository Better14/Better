// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"slices"
	"testing"
)

func sliceLen[T any](s []T) (int, bool) {
	return len(s), true
}

func TestChainWhereSelectToSlice(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	out := ToSlice(selectBySeq(whereSeq(From(nums), func(n int) bool { return n < 5 }), func(n int) int { return n + 1 }))
	want := []int{2, 3, 4, 5}
	if !slices.Equal(out, want) {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestChainFirst(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := firstFromSeq(selectBySeq(whereSeq(From(nums), func(n int) bool { return n%2 == 0 }), func(n int) int { return n * 2 }))
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestLazyPackageAPI(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := firstFromSeq(selectBySeq(whereSeq(From(nums), func(n int) bool { return n%2 == 0 }), func(n int) int { return n * 2 }))
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestLazyReceiverSelect(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	first := firstFromSeq(selectBySeq(whereSeq(From(nums), func(n int) bool { return n%2 == 0 }), func(n int) int { return n * 2 }))
	if first != 4 {
		t.Fatalf("got %v, want 4", first)
	}
}

func TestSumChain(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sum := sumSeq(whereSeq(From(nums), func(n int) bool { return n%2 == 0 }))
	if sum != 6 {
		t.Fatalf("sum: %v", sum)
	}
}

func TestRangeRepeatEmpty(t *testing.T) {
	got := toListSeq(Range(2, 3))
	want := []int{2, 3, 4}
	if !slices.Equal(got, want) {
		t.Fatalf("Range: got %v", got)
	}
	if firstFromSeq(Repeat(7, 2)) != 7 {
		t.Fatal("Repeat")
	}
	if anySeq(Empty[int](), func(int) bool { return true }) {
		t.Fatal("Empty")
	}
}

func TestSetOperations(t *testing.T) {
	a := []int{1, 2, 3, 4}
	b := []int{3, 4, 5, 6}
	union := toListSeq(unionSeq(From(a), From(b)))
	if len(union) != 6 {
		t.Fatalf("Union: %v", union)
	}
	except := toListSeq(exceptSeq(From(a), From(b)))
	if !slices.Equal(except, []int{1, 2}) {
		t.Fatalf("Except: %v", except)
	}
	inter := toListSeq(intersectSeq(From(a), From(b)))
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
	out := orderBySeq(From(people), func(p person) string { return p.name }).ThenBy(func(p person) int { return p.age }).ToSlice()
	if out[0].age != 20 || out[2].name != "bob" {
		t.Fatalf("Order/ThenBy: %v", out)
	}
}

func TestJoin(t *testing.T) {
	outer := []KeyValue[string, int]{{"a", 1}, {"b", 2}}
	inner := []KeyValue[string, string]{{"a", "x"}, {"b", "y"}}
	joined := toListSeq(joinSeq(
		From(outer),
		From(inner),
		func(kv KeyValue[string, int]) string { return kv.Key },
		func(kv KeyValue[string, string]) string { return kv.Key },
		func(o KeyValue[string, int], i KeyValue[string, string]) string {
			return o.Key + i.Value
		},
	))
	if !slices.Equal(joined, []string{"ax", "by"}) {
		t.Fatalf("Join: %v", joined)
	}
}

func TestFullJoin(t *testing.T) {
	type dept struct {
		id   int
		name string
	}
	type emp struct {
		id     int
		deptID int
		name   string
	}
	depts := []dept{{1, "sales"}, {2, "eng"}, {3, "hr"}}
	emps := []emp{{10, 1, "alice"}, {20, 2, "bob"}, {30, 99, "carol"}}

	joined := toListSeq(fullJoinSeq(
		From(depts),
		From(emps),
		func(d dept) int { return d.id },
		func(e emp) int { return e.deptID },
		func(d dept, e emp) string {
			return d.name + ":" + e.name
		},
		dept{},
		emp{},
	))

	want := []string{
		"sales:alice",
		"eng:bob",
		"hr:",
		":carol",
	}
	if !slices.Equal(joined, want) {
		t.Fatalf("FullJoin: got %v, want %v", joined, want)
	}
}

func TestMaterializers(t *testing.T) {
	nums := []int{1, 2, 2, 3}
	distinct := []int{1, 2, 3, 4}
	seq := From(nums)
	dict := toDictionarySeq(From(distinct),
		func(n int) int { return n },
		func(n int) string { return "v" },
	)
	if len(dict) != 4 {
		t.Fatalf("ToDictionary: %v", dict)
	}
	lookup := toLookupSeq(seq,
		func(n int) int { return n % 2 },
		func(n int) int { return n },
	)
	if lookup.Count() != 2 || len(lookup.Get(0)) != 2 {
		t.Fatalf("ToLookup: %v", lookup.Get(0))
	}
	set := toHashSetSeq(seq)
	if len(set) != 3 {
		t.Fatalf("ToHashSet: %v", set)
	}
}

func TestTryGetSeqLen(t *testing.T) {
	nums := []int{1, 2, 3}
	n, ok := sliceLen(nums)
	if !ok || n != 3 {
		t.Fatalf("got (%d, %v)", n, ok)
	}
	_, ok = tryGetSeqLenSeq(whereSeq(From([]int{1, 2, 3}), func(int) bool { return true }))
	if ok {
		t.Fatal("expected unknown count after Where")
	}
}

func TestTerminalOps(t *testing.T) {
	nums := []int{3, 1, 4, 1, 5}
	seq := From(nums)
	if maxSeq(seq) != 5 || minSeq(seq) != 1 {
		t.Fatal("Min/Max")
	}
	if averageSeq(seq) != 2.8 {
		t.Fatalf("Average: %v", averageSeq(seq))
	}
	if lastFromSeq(seq) != 5 || elementAtSeq(seq, 2) != 4 {
		t.Fatal("Last/ElementAt")
	}
	if singleSeq(whereSeq(From(nums), func(n int) bool { return n == 3 })) != 3 {
		t.Fatal("Single")
	}
}

func TestZipChunkAppendPrepend(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{10, 20, 30}
	zipped := toListSeq(zipSeq(From(a), From(b), func(x, y int) int { return x + y }))
	if !slices.Equal(zipped, []int{11, 22, 33}) {
		t.Fatalf("Zip: %v", zipped)
	}
	chunks := toListSeq(chunkSeq(From(a), 2))
	if len(chunks) != 2 || len(chunks[1]) != 1 {
		t.Fatalf("Chunk: %v", chunks)
	}
	appended := toListSeq(appendSeq(From(a), 4))
	if !slices.Equal(appended, []int{1, 2, 3, 4}) {
		t.Fatalf("Append: %v", appended)
	}
	prepended := toListSeq(prependSeq(From(a), 0))
	if !slices.Equal(prepended, []int{0, 1, 2, 3}) {
		t.Fatalf("Prepend: %v", prepended)
	}
}

func TestCastOfType(t *testing.T) {
	src := From([]any{1, "x", 2, 3.0})
	ints := toListSeq(ofTypeSeq[int](src))
	if !slices.Equal(ints, []int{1, 2}) {
		t.Fatalf("OfType: %v", ints)
	}
}

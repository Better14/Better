
package list

import (
	"iter"
	"slices"
	"testing"
)

func TestAppendDoublesCapacity(t *testing.T) {
	l := New[int]()
	for i := range 5 {
		l.Append(i)
	}
	if l.Len() != 5 {
		t.Fatalf("len: %d", l.Len())
	}
	if l.Cap() < 5 {
		t.Fatalf("cap: %d", l.Cap())
	}
	prevCap := l.Cap()
	for i := range 5 {
		l.Append(i + 10)
	}
	if l.Cap() < prevCap*2 {
		t.Fatalf("expected cap at least doubled from %d, got %d", prevCap, l.Cap())
	}
}

func TestAtSet(t *testing.T) {
	l := Of(10, 20, 30)
	if l.At(1) != 20 {
		t.Fatalf("got %d", l.At(1))
	}
	l.Set(0, 99)
	if l.At(0) != 99 {
		t.Fatalf("set failed: %d", l.At(0))
	}
}

func TestOfPreallocCap(t *testing.T) {
	l := Of(1, 2, 3, 4, 5)
	if l.Len() != 5 || l.Cap() < 5 {
		t.Fatalf("len=%d cap=%d", l.Len(), l.Cap())
	}
}

func TestAddRangeBatch(t *testing.T) {
	l := New[int]()
	l.AddRange(slices.Values([]int{1, 2, 3, 4, 5}))
	if !slices.Equal(l.ToSlice(), []int{1, 2, 3, 4, 5}) {
		t.Fatalf("got %v", l.ToSlice())
	}
}

func TestInsertAtEndUsesAppend(t *testing.T) {
	l := Of(1, 2)
	l.Insert(2, 3)
	if !slices.Equal(l.ToSlice(), []int{1, 2, 3}) {
		t.Fatalf("got %v", l.ToSlice())
	}
}

func TestAddRangeIndexOfContains(t *testing.T) {
	l := New[int]()
	l.AddRange(slices.Values([]int{1, 2, 3}))
	if !slices.Equal(l.ToSlice(), []int{1, 2, 3}) {
		t.Fatalf("got %v", l.ToSlice())
	}
	if IndexOf(l, 2) != 1 {
		t.Fatalf("index: %d", IndexOf(l, 2))
	}
	if IndexOf(l, 9) != -1 {
		t.Fatalf("index: %d", IndexOf(l, 9))
	}
	if !Contains(l, 3) || Contains(l, 0) {
		t.Fatalf("contains failed")
	}
}

func seq(vals ...int) iter.Seq[int] {
	return slices.Values(vals)
}

func TestAddRangeFromSeq(t *testing.T) {
	l := New[int]()
	l.AddRange(seq(4, 5))
	if got := l.ToSlice(); !slices.Equal(got, []int{4, 5}) {
		t.Fatalf("got %v", got)
	}
}

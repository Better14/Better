// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"cmp"
	"math/rand/v2"
	"slices"
)

// lazyAppend appends element to the end of l.
func lazyAppend[T any](l Lazy[T], element T) Lazy[T] {
	src := l.next
	appended := false
	return Lazy[T]{next: func() (T, bool) {
		v, ok := src()
		if ok {
			return v, true
		}
		if !appended {
			appended = true
			return element, true
		}
		var z T
		return z, false
	}}
}

// lazyPrepend prepends element to the start of l.
func lazyPrepend[T any](l Lazy[T], element T) Lazy[T] {
	src := l.next
	prefix := true
	return Lazy[T]{next: func() (T, bool) {
		if prefix {
			prefix = false
			return element, true
		}
		return src()
	}}
}

// lazyConcat concatenates l with other.
func lazyConcat[T any](l, other Lazy[T]) Lazy[T] {
	first := l.next
	second := other.next
	useFirst := true
	return Lazy[T]{next: func() (T, bool) {
		if useFirst {
			v, ok := first()
			if ok {
				return v, true
			}
			useFirst = false
		}
		return second()
	}}
}

// lazyZip pairs elements from l and other until either is exhausted.
func lazyZip[T, U, R any](l Lazy[T], other Lazy[U], fn func(T, U) R) Lazy[R] {
	a := l.next
	b := other.next
	return Lazy[R]{next: func() (R, bool) {
		va, oka := a()
		vb, okb := b()
		if !oka || !okb {
			var z R
			return z, false
		}
		return fn(va, vb), true
	}}
}

// lazySelectMany flattens each element to a sequence and concatenates.
func lazySelectMany[T, U any](l Lazy[T], fn func(T) Lazy[U]) Lazy[U] {
	src := l.next
	var inner Lazy[U]
	var innerNext func() (U, bool)
	return Lazy[U]{next: func() (U, bool) {
		for {
			if innerNext != nil {
				v, ok := innerNext()
				if ok {
					return v, true
				}
				innerNext = nil
			}
			v, ok := src()
			if !ok {
				var z U
				return z, false
			}
			inner = fn(v)
			innerNext = inner.next
		}
	}}
}

// lazySelectManySlice is SelectMany when fn returns a slice.
func lazySelectManySlice[T, U any](l Lazy[T], fn func(T) []U) Lazy[U] {
	return lazySelectMany(l, func(t T) Lazy[U] {
		return From(fn(t))
	})
}

// lazyReverse returns elements in reverse order (materializes l).
func lazyReverse[T any](l Lazy[T]) Lazy[T] {
	items := lazyToSlice(l)
	slices.Reverse(items)
	return From(items)
}

// lazySkipWhile skips elements while pred is true.
func lazySkipWhile[T any](l Lazy[T], pred func(T) bool) Lazy[T] {
	src := l.next
	skipping := true
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			if skipping && pred(v) {
				continue
			}
			skipping = false
			return v, true
		}
	}}
}

// lazyTakeWhile takes elements while pred is true.
func lazyTakeWhile[T any](l Lazy[T], pred func(T) bool) Lazy[T] {
	src := l.next
	done := false
	return Lazy[T]{next: func() (T, bool) {
		if done {
			var z T
			return z, false
		}
		v, ok := src()
		if !ok || !pred(v) {
			done = true
			if ok && !pred(v) {
				var z T
				return z, false
			}
			var z T
			return z, false
		}
		return v, true
	}}
}

// lazySkipLast skips the last count elements (materializes l).
func lazySkipLast[T any](l Lazy[T], count int) Lazy[T] {
	items := lazyToSlice(l)
	if count >= len(items) {
		return Empty[T]()
	}
	return From(items[:len(items)-count])
}

// lazyTakeLast takes the last count elements (materializes l).
func lazyTakeLast[T any](l Lazy[T], count int) Lazy[T] {
	items := lazyToSlice(l)
	if count <= 0 {
		return Empty[T]()
	}
	if count >= len(items) {
		return From(items)
	}
	return From(items[len(items)-count:])
}

// lazyChunk splits the sequence into chunks of size (materializes l).
func lazyChunk[T any](l Lazy[T], size int) Lazy[[]T] {
	if size <= 0 {
		panic("linq: chunk size must be positive")
	}
	items := lazyToSlice(l)
	i := 0
	return Lazy[[]T]{next: func() ([]T, bool) {
		if i >= len(items) {
			return nil, false
		}
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunk := slices.Clone(items[i:end])
		i = end
		return chunk, true
	}}
}

// lazyDefaultIfEmpty returns l, or a single defaultValue if l is empty.
func lazyDefaultIfEmpty[T any](l Lazy[T], defaultValue T) Lazy[T] {
	src := l.next
	v, ok := src()
	if ok {
		first := v
		started := false
		return Lazy[T]{next: func() (T, bool) {
			if !started {
				started = true
				return first, true
			}
			return src()
		}}
	}
	sent := false
	return Lazy[T]{next: func() (T, bool) {
		if sent {
			var z T
			return z, false
		}
		sent = true
		return defaultValue, true
	}}
}

// lazyIndex pairs each element with its index.
func lazyIndex[T any](l Lazy[T]) Lazy[Indexed[T]] {
	src := l.next
	i := 0
	return Lazy[Indexed[T]]{next: func() (Indexed[T], bool) {
		v, ok := src()
		if !ok {
			var z Indexed[T]
			return z, false
		}
		idx := Indexed[T]{Index: i, Value: v}
		i++
		return idx, true
	}}
}

// lazyCast casts each element to T (panics on failed cast).
func lazyCast[T any](l Lazy[any]) Lazy[T] {
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		v, ok := src()
		if !ok {
			var z T
			return z, false
		}
		t, ok := v.(T)
		if !ok {
			panic("linq: cast failed")
		}
		return t, true
	}}
}

// lazyOfType filters elements assignable to T.
func lazyOfType[T any](l Lazy[any]) Lazy[T] {
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			t, ok := v.(T)
			if ok {
				return t, true
			}
		}
	}}
}

// lazyShuffle returns elements in random order (materializes l).
func lazyShuffle[T any](l Lazy[T]) Lazy[T] {
	items := lazyToSlice(l)
	rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	return From(items)
}

// lazyDistinctBy returns distinct elements by key.
func lazyDistinctBy[T any, K comparable](l Lazy[T], keyFn func(T) K) Lazy[T] {
	seen := make(map[K]struct{})
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			k := keyFn(v)
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			return v, true
		}
	}}
}

// lazyExcept returns elements in l not present in other (comparable T).
func lazyExcept[T comparable](l, other Lazy[T]) Lazy[T] {
	exclude := make(map[T]struct{})
	for v, ok := other.next(); ok; v, ok = other.next() {
		exclude[v] = struct{}{}
	}
	return lazyWhere(l, func(v T) bool {
		_, found := exclude[v]
		return !found
	})
}

// lazyExceptBy returns elements in l whose key is not in other.
func lazyExceptBy[T any, K comparable](l, other Lazy[T], keyFn func(T) K) Lazy[T] {
	exclude := make(map[K]struct{})
	for v, ok := other.next(); ok; v, ok = other.next() {
		exclude[keyFn(v)] = struct{}{}
	}
	return lazyWhere(l, func(v T) bool {
		_, found := exclude[keyFn(v)]
		return !found
	})
}

// lazyIntersect returns elements present in both sequences.
func lazyIntersect[T comparable](l, other Lazy[T]) Lazy[T] {
	inOther := make(map[T]struct{})
	for v, ok := other.next(); ok; v, ok = other.next() {
		inOther[v] = struct{}{}
	}
	seen := make(map[T]struct{})
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			if _, ok := inOther[v]; !ok {
				continue
			}
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			return v, true
		}
	}}
}

// lazyIntersectBy returns elements whose key appears in both sequences.
func lazyIntersectBy[T any, K comparable](l, other Lazy[T], keyFn func(T) K) Lazy[T] {
	inOther := make(map[K]struct{})
	for v, ok := other.next(); ok; v, ok = other.next() {
		inOther[keyFn(v)] = struct{}{}
	}
	seen := make(map[K]struct{})
	src := l.next
	return Lazy[T]{next: func() (T, bool) {
		for {
			v, ok := src()
			if !ok {
				var z T
				return z, false
			}
			k := keyFn(v)
			if _, ok := inOther[k]; !ok {
				continue
			}
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			return v, true
		}
	}}
}

// lazyUnion returns distinct elements from l and other.
func lazyUnion[T comparable](l, other Lazy[T]) Lazy[T] {
	return lazyDistinct(lazyConcat(l, other))
}

// lazyUnionBy returns distinct elements from l and other by key.
func lazyUnionBy[T any, K comparable](l, other Lazy[T], keyFn func(T) K) Lazy[T] {
	return lazyDistinctBy(lazyConcat(l, other), keyFn)
}

// lazyContains reports whether l contains value.
func lazyContains[T comparable](l Lazy[T], value T) bool {
	for v, ok := l.next(); ok; v, ok = l.next() {
		if v == value {
			return true
		}
	}
	return false
}

// lazySequenceEqual reports whether l and other contain equal elements in order.
func lazySequenceEqual[T comparable](l, other Lazy[T]) bool {
	a := l.next
	b := other.next
	for {
		va, oka := a()
		vb, okb := b()
		if oka != okb {
			return false
		}
		if !oka {
			return true
		}
		if va != vb {
			return false
		}
	}
}

// lazyElementAt returns the element at index, or panics.
func lazyElementAt[T any](l Lazy[T], index int) T {
	if index < 0 {
		panic("linq: index out of range")
	}
	i := 0
	for v, ok := l.next(); ok; v, ok = l.next() {
		if i == index {
			return v
		}
		i++
	}
	panic("linq: index out of range")
}

// lazyElementAtOrDefault returns the element at index or defaultValue.
func lazyElementAtOrDefault[T any](l Lazy[T], index int, defaultValue T) T {
	if index < 0 {
		return defaultValue
	}
	i := 0
	for v, ok := l.next(); ok; v, ok = l.next() {
		if i == index {
			return v
		}
		i++
	}
	return defaultValue
}

// lazyLast returns the last element, or panics if empty.
func lazyLast[T any](l Lazy[T]) T {
	v, ok := lazyLastValue(l)
	if !ok {
		panic("linq: sequence contains no elements")
	}
	return v
}

// lazyLastValue returns the last element and whether it exists.
func lazyLastValue[T any](l Lazy[T]) (T, bool) {
	var last T
	found := false
	for v, ok := l.next(); ok; v, ok = l.next() {
		last = v
		found = true
	}
	return last, found
}

// lazyLastOrDefault returns the last element or the zero value.
func lazyLastOrDefault[T any](l Lazy[T]) T {
	v, ok := lazyLastValue(l)
	if ok {
		return v
	}
	var z T
	return z
}

// lazySingle returns the only element, or panics.
func lazySingle[T any](l Lazy[T]) T {
	v, ok := l.next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	if _, ok = l.next(); ok {
		panic("linq: sequence contains more than one element")
	}
	return v
}

// lazySingleOrDefault returns the only element or defaultValue; panics if more than one.
func lazySingleOrDefault[T any](l Lazy[T], defaultValue T) T {
	v, ok := l.next()
	if !ok {
		return defaultValue
	}
	if _, ok = l.next(); ok {
		panic("linq: sequence contains more than one element")
	}
	return v
}

// lazyMax returns the maximum element.
func lazyMax[T cmp.Ordered](l Lazy[T]) T {
	v, ok := l.next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	max := v
	for v, ok = l.next(); ok; v, ok = l.next() {
		if v > max {
			max = v
		}
	}
	return max
}

// lazyMaxBy returns the element with the maximum key.
func lazyMaxBy[T any, K cmp.Ordered](l Lazy[T], key func(T) K) T {
	v, ok := l.next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	best := v
	bestKey := key(v)
	for v, ok = l.next(); ok; v, ok = l.next() {
		k := key(v)
		if k > bestKey {
			best = v
			bestKey = k
		}
	}
	return best
}

// lazyMin returns the minimum element.
func lazyMin[T cmp.Ordered](l Lazy[T]) T {
	v, ok := l.next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	min := v
	for v, ok = l.next(); ok; v, ok = l.next() {
		if v < min {
			min = v
		}
	}
	return min
}

// lazyMinBy returns the element with the minimum key.
func lazyMinBy[T any, K cmp.Ordered](l Lazy[T], key func(T) K) T {
	v, ok := l.next()
	if !ok {
		panic("linq: sequence contains no elements")
	}
	best := v
	bestKey := key(v)
	for v, ok = l.next(); ok; v, ok = l.next() {
		k := key(v)
		if k < bestKey {
			best = v
			bestKey = k
		}
	}
	return best
}

// lazyAverage returns the arithmetic mean as float64.
func lazyAverage[U Number](l Lazy[U]) float64 {
	var sum float64
	n := 0
	for v, ok := l.next(); ok; v, ok = l.next() {
		sum += float64(v)
		n++
	}
	if n == 0 {
		panic("linq: sequence contains no elements")
	}
	return sum / float64(n)
}

// lazyLongCount returns the number of elements as int64.
func lazyLongCount[T any](l Lazy[T]) int64 {
	var n int64
	for _, ok := l.next(); ok; _, ok = l.next() {
		n++
	}
	return n
}

// lazyCountBy counts elements by key.
func lazyCountBy[T any, K comparable](l Lazy[T], keyFn func(T) K) Lazy[KeyValue[K, int]] {
	counts := map[K]int{}
	var keys []K
	for v, ok := l.next(); ok; v, ok = l.next() {
		k := keyFn(v)
		if _, seen := counts[k]; !seen {
			keys = append(keys, k)
		}
		counts[k]++
	}
	i := 0
	return Lazy[KeyValue[K, int]]{next: func() (KeyValue[K, int], bool) {
		if i >= len(keys) {
			var z KeyValue[K, int]
			return z, false
		}
		k := keys[i]
		i++
		return KeyValue[K, int]{Key: k, Value: counts[k]}, true
	}}
}

// lazyAggregateWithSeed folds l starting with seed.
func lazyAggregateWithSeed[T any, U any](l Lazy[T], seed U, fn func(U, T) U) U {
	acc := seed
	for v, ok := l.next(); ok; v, ok = l.next() {
		acc = fn(acc, v)
	}
	return acc
}

// lazyAggregateBy groups by key and aggregates each group.
func lazyAggregateBy[T any, K comparable, A any, R any](l Lazy[T], keyFn func(T) K, seed A, fn func(A, T) A, resultFn func(A) R) Lazy[KeyValue[K, R]] {
	groups := map[K]A{}
	var keys []K
	for v, ok := l.next(); ok; v, ok = l.next() {
		k := keyFn(v)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k)
			groups[k] = seed
		}
		groups[k] = fn(groups[k], v)
	}
	i := 0
	return Lazy[KeyValue[K, R]]{next: func() (KeyValue[K, R], bool) {
		if i >= len(keys) {
			var z KeyValue[K, R]
			return z, false
		}
		k := keys[i]
		i++
		return KeyValue[K, R]{Key: k, Value: resultFn(groups[k])}, true
	}}
}

// TryGetNonEnumeratedCount reports a known length without enumerating.
func tryGetNonEnumeratedCount[T any](l Lazy[T]) (int, bool) {
	if l.hasKnownLen {
		return l.knownLen, true
	}
	return 0, false
}

// lazyToArray is an alias for materializing to a slice.
func lazyToArray[T any](l Lazy[T]) []T {
	return lazyToSlice(l)
}

// lazyToHashSet materializes comparable elements into a set.
func lazyToHashSet[T comparable](l Lazy[T]) HashSet[T] {
	set := make(HashSet[T])
	for v, ok := l.next(); ok; v, ok = l.next() {
		set[v] = struct{}{}
	}
	return set
}

// lazyToDictionary builds a map from key selector; panics on duplicate keys.
func lazyToDictionary[T any, K comparable, V any](l Lazy[T], keyFn func(T) K, valueFn func(T) V) map[K]V {
	out := make(map[K]V)
	for v, ok := l.next(); ok; v, ok = l.next() {
		k := keyFn(v)
		if _, dup := out[k]; dup {
			panic("linq: duplicate key in ToDictionary")
		}
		out[k] = valueFn(v)
	}
	return out
}

// lazyToLookup builds a Lookup from key selector.
func lazyToLookup[T any, K comparable, V any](l Lazy[T], keyFn func(T) K, valueFn func(T) V) Lookup[K, V] {
	groups := map[K][]V{}
	var keys []K
	for item, ok := l.next(); ok; item, ok = l.next() {
		k := keyFn(item)
		if _, seen := groups[k]; !seen {
			keys = append(keys, k)
			groups[k] = nil
		}
		groups[k] = append(groups[k], valueFn(item))
	}
	return Lookup[K, V]{groups: groups, keys: keys}
}

// asEnumerable returns l unchanged (identity for slices converted to lazy).
func asEnumerable[T any](l Lazy[T]) Lazy[T] {
	return l
}

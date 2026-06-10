// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

// lazyJoin inner-joins l with inner on outerKey and innerKey.
func lazyJoin[T, U, K comparable, R any](
	outer Lazy[T],
	inner Lazy[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
) Lazy[R] {
	lookup := lazyToLookup(inner, innerKey, func(u U) U { return u })
	idx := 0
	var outerVal T
	var innerVals []U
	innerIdx := 0
	advanceOuter := func() bool {
		for {
			v, ok := outer.next()
			if !ok {
				return false
			}
			outerVal = v
			innerVals = lookup.Get(outerKey(v))
			innerIdx = 0
			if len(innerVals) > 0 {
				return true
			}
		}
	}
	hasOuter := advanceOuter()
	return Lazy[R]{next: func() (R, bool) {
		for {
			if !hasOuter {
				var z R
				return z, false
			}
			if innerIdx < len(innerVals) {
				r := resultFn(outerVal, innerVals[innerIdx])
				innerIdx++
				return r, true
			}
			hasOuter = advanceOuter()
		}
	}}
}

// lazyGroupJoin groups inner by innerKey and joins with outer.
func lazyGroupJoin[T, U, K comparable, R any](
	outer Lazy[T],
	inner Lazy[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, Lazy[U]) R,
) Lazy[R] {
	lookup := lazyToLookup(inner, innerKey, func(u U) U { return u })
	return lazySelectBy(outer, func(t T) R {
		group := From(lookup.Get(outerKey(t)))
		return resultFn(t, group)
	})
}

// lazyLeftJoin left-joins outer with inner; defaultInner used when no match.
func lazyLeftJoin[T, U, K comparable, R any](
	outer Lazy[T],
	inner Lazy[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
	defaultInner U,
) Lazy[R] {
	lookup := lazyToLookup(inner, innerKey, func(u U) U { return u })
	return lazySelectBy(outer, func(t T) R {
		vals := lookup.Get(outerKey(t))
		if len(vals) == 0 {
			return resultFn(t, defaultInner)
		}
		return resultFn(t, vals[0])
	})
}

// lazyRightJoin right-joins outer with inner; defaultOuter used when no match.
func lazyRightJoin[T, U, K comparable, R any](
	outer Lazy[T],
	inner Lazy[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
	defaultOuter T,
) Lazy[R] {
	outerLookup := lazyToLookup(outer, outerKey, func(t T) T { return t })
	return lazySelectManyBy(inner, func(u U) Lazy[R] {
		outers := outerLookup.Get(k)
		if len(outers) == 0 {
			return From([]R{resultFn(defaultOuter, u)})
		}
		out := make([]R, len(outers))
		for i, t := range outers {
			out[i] = resultFn(t, u)
		}
		return From(out)
	})
}

// lazySelectManyBy is SelectMany with a result lazy per element.
func lazySelectManyBy[T, R any](l Lazy[T], fn func(T) Lazy[R]) Lazy[R] {
	return lazySelectMany(l, fn)
}

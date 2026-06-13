// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linq

import (
	"iter"
	"slices"
)

func joinSeq[T, U any, K comparable, R any](
	outer iter.Seq[T],
	inner iter.Seq[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
) iter.Seq[R] {
	lookup := toLookupSeq(inner, innerKey, func(u U) U { return u })
	return func(yield func(R) bool) {
		for outerVal := range outer {
			innerVals := lookup.Get(outerKey(outerVal))
			for _, innerVal := range innerVals {
				if !yield(resultFn(outerVal, innerVal)) {
					return
				}
			}
		}
	}
}

func groupJoinSeq[T, U any, K comparable, R any](
	outer iter.Seq[T],
	inner iter.Seq[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, iter.Seq[U]) R,
) iter.Seq[R] {
	lookup := toLookupSeq(inner, innerKey, func(u U) U { return u })
	return selectBySeq(outer, func(t T) R {
		group := slices.Values(lookup.Get(outerKey(t)))
		return resultFn(t, group)
	})
}

func leftJoinSeq[T, U any, K comparable, R any](
	outer iter.Seq[T],
	inner iter.Seq[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
	defaultInner U,
) iter.Seq[R] {
	lookup := toLookupSeq(inner, innerKey, func(u U) U { return u })
	return selectBySeq(outer, func(t T) R {
		vals := lookup.Get(outerKey(t))
		if len(vals) == 0 {
			return resultFn(t, defaultInner)
		}
		return resultFn(t, vals[0])
	})
}

func rightJoinSeq[T, U any, K comparable, R any](
	outer iter.Seq[T],
	inner iter.Seq[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
	defaultOuter T,
) iter.Seq[R] {
	outerLookup := toLookupSeq(outer, outerKey, func(t T) T { return t })
	return selectManySeq(inner, func(u U) iter.Seq[R] {
		outers := outerLookup.Get(innerKey(u))
		if len(outers) == 0 {
			return slices.Values([]R{resultFn(defaultOuter, u)})
		}
		out := make([]R, len(outers))
		for i, t := range outers {
			out[i] = resultFn(t, u)
		}
		return slices.Values(out)
	})
}

func fullJoinSeq[T, U any, K comparable, R any](
	outer iter.Seq[T],
	inner iter.Seq[U],
	outerKey func(T) K,
	innerKey func(U) K,
	resultFn func(T, U) R,
	defaultOuter T,
	defaultInner U,
) iter.Seq[R] {
	innerLookup := toLookupSeq(inner, innerKey, func(u U) U { return u })
	matchedKeys := make(map[K]struct{})
	return func(yield func(R) bool) {
		for o := range outer {
			k := outerKey(o)
			matchedKeys[k] = struct{}{}
			inners := innerLookup.Get(k)
			if len(inners) == 0 {
				if !yield(resultFn(o, defaultInner)) {
					return
				}
				continue
			}
			for _, i := range inners {
				if !yield(resultFn(o, i)) {
					return
				}
			}
		}
		for i := range inner {
			k := innerKey(i)
			if _, ok := matchedKeys[k]; ok {
				continue
			}
			if !yield(resultFn(defaultOuter, i)) {
				return
			}
		}
	}
}

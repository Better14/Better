// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package testerrors provides helpers for comparing structured errors in tests.
package testerrors

import (
	"errors"
	"reflect"
)

// IsNil reports whether err is nil or an interface holding a nil pointer.
func IsNil(err error) bool {
	if err == nil {
		return true
	}
	v := reflect.ValueOf(err)
	return v.Kind() == reflect.Pointer && v.IsNil()
}

// EqualValues reports whether x and y are deeply equal, except that
// errors.Base.StackTrace fields are ignored.
func EqualValues(x, y any) bool {
	return equalValues(reflect.ValueOf(x), reflect.ValueOf(y))
}

func equalValues(vx, vy reflect.Value) bool {
	if !vx.IsValid() || !vy.IsValid() {
		return vx.IsValid() == vy.IsValid()
	}
	if vx.Type() != vy.Type() {
		return false
	}

	switch vx.Kind() {
	case reflect.Pointer:
		if vx.IsNil() || vy.IsNil() {
			return vx.IsNil() && vy.IsNil()
		}
		return equalValues(vx.Elem(), vy.Elem())
	case reflect.Interface:
		if vx.IsNil() || vy.IsNil() {
			return vx.IsNil() && vy.IsNil()
		}
		return equalValues(vx.Elem(), vy.Elem())
	case reflect.Struct:
		if vx.Type() == reflect.TypeOf(errors.Base{}) {
			return infoDataEqual(vx.Interface().(errors.Base), vy.Interface().(errors.Base))
		}
		for i := 0; i < vx.NumField(); i++ {
			sf := vx.Type().Field(i)
			if !sf.IsExported() {
				continue
			}
			if sf.Type == reflect.TypeOf(errors.Base{}) {
				if !infoDataEqual(vx.Field(i).Interface().(errors.Base), vy.Field(i).Interface().(errors.Base)) {
					return false
				}
				continue
			}
			if !equalValues(vx.Field(i), vy.Field(i)) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(vx.Interface(), vy.Interface())
	}
}

func infoDataEqual(a, b errors.Base) bool {
	// Ignore Message and StackTrace; callers compare Error() strings when needed.
	if (a.InnerError == nil) != (b.InnerError == nil) {
		return false
	}
	if a.InnerError != nil && !infoDataEqual(*a.InnerError, *b.InnerError) {
		return false
	}
	return true
}

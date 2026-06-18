// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.jsonv2

package json

import (
	"errors"
	"reflect"
)

func newSyntaxError(msg string, offset int64) *SyntaxError {
	e := &SyntaxError{msg: msg, Offset: offset}
	errors.InitCustom(&e.Error, "%s", e.Error())
	return e
}

func newUnmarshalTypeError(value string, typ reflect.Type, offset int64) *UnmarshalTypeError {
	e := &UnmarshalTypeError{Value: value, Type: typ, Offset: offset}
	errors.InitCustom(&e.Error, "%s", e.Error())
	return e
}

func newInvalidUnmarshalError(typ reflect.Type) *InvalidUnmarshalError {
	e := &InvalidUnmarshalError{Type: typ}
	errors.InitCustom(&e.Error, "%s", e.Error())
	return e
}

func newMarshalerError(typ reflect.Type, err error, sourceFunc string) *MarshalerError {
	e := &MarshalerError{Type: typ, Err: err, sourceFunc: sourceFunc}
	errors.InitCustom(&e.Error, "%s", e.Error())
	return e
}

func newUnsupportedTypeError(typ reflect.Type) *UnsupportedTypeError {
	e := &UnsupportedTypeError{Type: typ}
	errors.InitCustom(&e.Error, "%s", e.Error())
	return e
}

func newUnsupportedValueError(v reflect.Value, str string) *UnsupportedValueError {
	e := &UnsupportedValueError{Value: v, Str: str}
	errors.InitCustom(&e.Error, "%s", e.Error())
	return e
}

func newJsonError(err error) jsonError {
	je := jsonError{err: err}
	errors.InitCustom(&je.Error, "%s", err.Error())
	return je
}

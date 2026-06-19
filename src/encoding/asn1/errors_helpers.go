// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asn1

import (
	"errors"
	"reflect"
)

func structuralErrorMessage(msg string) string {
	return "asn1: structure error: " + msg
}

func newStructuralError(msg string) StructuralError {
	e := StructuralError{Msg: msg}
	errors.InitCustom(&e.Base, "%s", structuralErrorMessage(msg))
	return e
}

func syntaxErrorMessage(msg string) string {
	return "asn1: syntax error: " + msg
}

func newSyntaxError(msg string) SyntaxError {
	e := SyntaxError{Msg: msg}
	errors.InitCustom(&e.Base, "%s", syntaxErrorMessage(msg))
	return e
}

func invalidUnmarshalErrorMessage(typ reflect.Type) string {
	if typ == nil {
		return "asn1: Unmarshal recipient value is nil"
	}
	if typ.Kind() != reflect.Pointer {
		return "asn1: Unmarshal recipient value is non-pointer " + typ.String()
	}
	return "asn1: Unmarshal recipient value is nil " + typ.String()
}

func newInvalidUnmarshalError(typ reflect.Type) *invalidUnmarshalError {
	e := &invalidUnmarshalError{Type: typ}
	errors.InitCustom(&e.Base, "%s", invalidUnmarshalErrorMessage(typ))
	return e
}

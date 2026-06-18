// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http2

import (
	"errors"
	"fmt"
	"reflect"
)

// An ErrCode is an unsigned 32-bit error code as defined in the HTTP/2 spec.
type ErrCode uint32

const (
	ErrCodeNo                 ErrCode = 0x0
	ErrCodeProtocol           ErrCode = 0x1
	ErrCodeInternal           ErrCode = 0x2
	ErrCodeFlowControl        ErrCode = 0x3
	ErrCodeSettingsTimeout    ErrCode = 0x4
	ErrCodeStreamClosed       ErrCode = 0x5
	ErrCodeFrameSize          ErrCode = 0x6
	ErrCodeRefusedStream      ErrCode = 0x7
	ErrCodeCancel             ErrCode = 0x8
	ErrCodeCompression        ErrCode = 0x9
	ErrCodeConnect            ErrCode = 0xa
	ErrCodeEnhanceYourCalm    ErrCode = 0xb
	ErrCodeInadequateSecurity ErrCode = 0xc
	ErrCodeHTTP11Required     ErrCode = 0xd
)

var errCodeName = map[ErrCode]string{
	ErrCodeNo:                 "NO_ERROR",
	ErrCodeProtocol:           "PROTOCOL_ERROR",
	ErrCodeInternal:           "INTERNAL_ERROR",
	ErrCodeFlowControl:        "FLOW_CONTROL_ERROR",
	ErrCodeSettingsTimeout:    "SETTINGS_TIMEOUT",
	ErrCodeStreamClosed:       "STREAM_CLOSED",
	ErrCodeFrameSize:          "FRAME_SIZE_ERROR",
	ErrCodeRefusedStream:      "REFUSED_STREAM",
	ErrCodeCancel:             "CANCEL",
	ErrCodeCompression:        "COMPRESSION_ERROR",
	ErrCodeConnect:            "CONNECT_ERROR",
	ErrCodeEnhanceYourCalm:    "ENHANCE_YOUR_CALM",
	ErrCodeInadequateSecurity: "INADEQUATE_SECURITY",
	ErrCodeHTTP11Required:     "HTTP_1_1_REQUIRED",
}

func (e ErrCode) String() string {
	if s, ok := errCodeName[e]; ok {
		return s
	}
	return fmt.Sprintf("unknown error code 0x%x", uint32(e))
}

func (e ErrCode) stringToken() string {
	if s, ok := errCodeName[e]; ok {
		return s
	}
	return fmt.Sprintf("ERR_UNKNOWN_%d", uint32(e))
}

// ConnectionError is an error that results in the termination of the
// entire connection.
type ConnectionError struct {
	errors.Error
	Code ErrCode
}

func connectionErrorMessage(code ErrCode) string {
	return fmt.Sprintf("connection error: %s", code)
}

// NewConnectionError returns a ConnectionError with a stack trace captured at the call site.
func NewConnectionError(code ErrCode) ConnectionError {
	e := ConnectionError{Code: code}
	errors.InitCustom(&e.Error, "%s", connectionErrorMessage(code))
	return e
}

func (e ConnectionError) Error() string { return connectionErrorMessage(e.Code) }

// StreamError is an error that only affects one stream within an
// HTTP/2 connection.
type StreamError struct {
	errors.Error
	StreamID uint32
	Code     ErrCode
	Cause    error // optional additional detail
}

// errFromPeer is a sentinel error value for StreamError.Cause to
// indicate that the StreamError was sent from the peer over the wire
// and wasn't locally generated in the Transport.
var errFromPeer = errors.New("received from peer")

func streamErrorMessage(id uint32, code ErrCode, cause error) string {
	if cause != nil {
		return fmt.Sprintf("stream error: stream ID %d; %v; %v", id, code, cause)
	}
	return fmt.Sprintf("stream error: stream ID %d; %v", id, code)
}

func streamError(id uint32, code ErrCode) StreamError {
	return NewStreamError(id, code, nil)
}

// NewStreamError returns a StreamError with a stack trace captured at the call site.
func NewStreamError(id uint32, code ErrCode, cause error) StreamError {
	e := StreamError{StreamID: id, Code: code, Cause: cause}
	errors.InitCustom(&e.Error, "%s", streamErrorMessage(id, code, cause))
	return e
}

func (e StreamError) Error() string {
	return streamErrorMessage(e.StreamID, e.Code, e.Cause)
}

// This As function permits converting a StreamError into a x/net/http2.StreamError.
func (e StreamError) As(target any) bool {
	dst := reflect.ValueOf(target).Elem()
	dstType := dst.Type()
	if dstType.Kind() != reflect.Struct {
		return false
	}
	src := reflect.ValueOf(e)
	srcType := src.Type()
	for i := range dstType.NumField() {
		df := dstType.Field(i)
		sf, ok := srcType.FieldByName(df.Name)
		if !ok || !src.FieldByIndex(sf.Index).Type().ConvertibleTo(df.Type) {
			return false
		}
	}
	for i := range dstType.NumField() {
		df := dstType.Field(i)
		dst.Field(i).Set(src.FieldByName(df.Name).Convert(df.Type))
	}
	return true
}

// 6.9.1 The Flow Control Window
// "If a sender receives a WINDOW_UPDATE that causes a flow control
// window to exceed this maximum it MUST terminate either the stream
// or the connection, as appropriate. For streams, [...]; for the
// connection, a GOAWAY frame with a FLOW_CONTROL_ERROR code."
type goAwayFlowError struct {
	errors.Error
}

func newGoAwayFlowError() goAwayFlowError {
	e := goAwayFlowError{}
	errors.InitCustom(&e.Error, "connection exceeded flow control window size")
	return e
}

func (goAwayFlowError) Error() string { return "connection exceeded flow control window size" }

// connError represents an HTTP/2 ConnectionError error code, along
// with a string (for debugging) explaining why.
//
// Errors of this type are only returned by the frame parser functions
// and converted into ConnectionError(Code), after stashing away
// the Reason into the Framer's errDetail field, accessible via
// the (*Framer).ErrorDetail method.
type connError struct {
	errors.Error
	Code   ErrCode // the ConnectionError error code
	Reason string  // additional reason
}

func connErrorMessage(code ErrCode, reason string) string {
	return fmt.Sprintf("http2: connection error: %v: %v", code, reason)
}

func newConnError(code ErrCode, reason string) connError {
	e := connError{Code: code, Reason: reason}
	errors.InitCustom(&e.Error, "%s", connErrorMessage(code, reason))
	return e
}

func (e connError) Error() string {
	return connErrorMessage(e.Code, e.Reason)
}

type pseudoHeaderError struct {
	errors.Error
	name string
}

func newPseudoHeaderError(name string) pseudoHeaderError {
	e := pseudoHeaderError{name: name}
	errors.InitCustom(&e.Error, "invalid pseudo-header %q", name)
	return e
}

func (e pseudoHeaderError) Error() string {
	return fmt.Sprintf("invalid pseudo-header %q", e.name)
}

type duplicatePseudoHeaderError struct {
	errors.Error
	name string
}

func newDuplicatePseudoHeaderError(name string) duplicatePseudoHeaderError {
	e := duplicatePseudoHeaderError{name: name}
	errors.InitCustom(&e.Error, "duplicate pseudo-header %q", name)
	return e
}

func (e duplicatePseudoHeaderError) Error() string {
	return fmt.Sprintf("duplicate pseudo-header %q", e.name)
}

type headerFieldNameError struct {
	errors.Error
	name string
}

func newHeaderFieldNameError(name string) headerFieldNameError {
	e := headerFieldNameError{name: name}
	errors.InitCustom(&e.Error, "invalid header field name %q", name)
	return e
}

func (e headerFieldNameError) Error() string {
	return fmt.Sprintf("invalid header field name %q", e.name)
}

type headerFieldValueError struct {
	errors.Error
	name string
}

func newHeaderFieldValueError(name string) headerFieldValueError {
	e := headerFieldValueError{name: name}
	errors.InitCustom(&e.Error, "invalid header field value for %q", name)
	return e
}

func (e headerFieldValueError) Error() string {
	return fmt.Sprintf("invalid header field value for %q", e.name)
}

var (
	errMixPseudoHeaderTypes = errors.New("mix of request and response pseudo headers")
	errPseudoAfterRegular   = errors.New("pseudo header field after regular")
)

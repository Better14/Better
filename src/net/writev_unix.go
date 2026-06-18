// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package net

import (
	"runtime"
	"syscall"
)

func (c *conn) writeBuffers(v *Buffers) (int64, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	n, err := c.fd.writeBuffers(v)
	if err != nil {
		return n, NewOpError("writev", c.fd.net, c.fd.laddr, c.fd.raddr, err)
	}
	return n, nil
}

func (fd *netFD) writeBuffers(v *Buffers) (n int64, err error) {
	n, err = fd.pfd.Writev((*[][]byte)(v))
	runtime.KeepAlive(fd)
	return n, wrapSyscallError("writev", err)
}

// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test ! with T! helpers, n-tuple returns, and error-only functions.

package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

func readAllLimited(r io.Reader, limit int64) []byte! {
	return io.ReadAll(io.LimitReader(r, limit))!
}

func decodeImage(data []byte) image.Image! {
	return image.Decode(bytes.NewReader(data))!
}

func saveFile(path string) error {
	f := os.Create(path)!
	defer f.Close()
	_, err := f.WriteString("ok")
	return err
}

func load() []byte! {
	return readAllLimited(bytes.NewReader([]byte("hi")), 10)!
}

func run() error {
	data := load()!
	_ = data
	return saveFile(os.TempDir() + string(os.PathSeparator) + "result_force_helpers.txt")
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
}

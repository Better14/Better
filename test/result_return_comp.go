// run

// Copyright authors of this Go fork
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type BenchmarkDB struct {
	URL     string
	Cleanup func()
}

func parseURL(raw string) *int! {
	v := fakeParse(raw)!
	return v
}

func fakeParse(raw string) (*int, error) {
	if raw == "fail" {
		return nil, fmt.Errorf("fail")
	}
	n := 1
	return &n, nil
}

func CreateBenchmarkDB() BenchmarkDB! {
	base := "ok"
	parsed := parseURL(base)!
	_ = parsed
	return BenchmarkDB{URL: "bench", Cleanup: nil}
}

func CreateBenchmarkDBFail() BenchmarkDB! {
	parseURL("fail")!
	return BenchmarkDB{URL: "bench", Cleanup: nil}
}

func main() {
	v, err := CreateBenchmarkDB()
	if err != nil || v.URL != "bench" {
		panic(fmt.Sprintf("got %+v err=%v", v, err))
	}
	v, err = CreateBenchmarkDBFail()
	if err == nil || v.URL != "" {
		panic(fmt.Sprintf("fail path: got %+v err=%v", v, err))
	}
	fmt.Println("ok")
}

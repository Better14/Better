// run

package main

import (
	"errors"
	"strings"
)

func expand(s string) ([]string, error) {
	var endpoints []string
	for endpoint in strings.SplitSeq(s, ",") {
		_, err := strings.NewReader(endpoint).Read(make([]byte, 0))
		if err != nil {
			return nil, errors.New("kms: invalid endpoint '{endpoint}': {err}")
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func main() {
	if _, err := expand("a,b"); err != nil {
		_ = err
	}
}

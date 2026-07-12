// run

package main

import (
	"errors"
	"strings"
)

func expand(s string) []string! {
	var endpoints []string
	for endpoint := range strings.SplitSeq(s, ",") {
		endpoint = strings.TrimSpace(endpoint)
		pattern, err := find(endpoint)
		if err != nil {
			return errors.New("invalid endpoint '%s': %v", endpoint, err)
		}
		endpoints = append(endpoints, pattern)
	}
	return endpoints
}

func find(s string) (string, error) {
	if s == "" {
		return "", errors.New("empty")
	}
	return s, nil
}

func main() {
	if _, err := expand("a,b"); err != nil {
		panic(err)
	}
}

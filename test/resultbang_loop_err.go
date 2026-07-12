// run

package main

import "errors"

func expand(s string) []string! {
	var endpoints []string
	for _, endpoint := range []string{s} {
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
	if _, err := expand("x"); err != nil {
		panic(err)
	}
}

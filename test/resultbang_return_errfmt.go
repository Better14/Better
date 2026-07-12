// run

package main

import "errors"

func expand(s string) []string! {
	var endpoints []string
	pattern, err := find(s)
	if err != nil {
		return errors.New("invalid endpoint '%s': %v", s, err)
	}
	endpoints = append(endpoints, pattern)
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

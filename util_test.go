package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestTrim(t *testing.T) {
	var a = "/tmp/blog/"
	a = strings.TrimSuffix(a, "/")
	fmt.Println(a)
}

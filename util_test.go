package main

import (
	"strings"
	"testing"
)

func TestTrim(t *testing.T) {
	var a = "/tmp/blog/"
	a = strings.TrimSuffix(a, "/")
	if a != "/tmp/blog" {
		t.Fatalf("trimmed path = %q, want /tmp/blog", a)
	}
}

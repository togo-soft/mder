package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunInitRejectsEmptyName(t *testing.T) {
	t.Parallel()

	err := runInit("")
	if err == nil {
		t.Fatal("expected empty name to fail")
	}
	if !strings.Contains(err.Error(), "folder name empty") {
		t.Fatalf("error = %q, want folder name empty", err)
	}
}

func TestRunInitRejectsInvalidName(t *testing.T) {
	t.Parallel()

	err := runInit("bad-name")
	if err == nil {
		t.Fatal("expected invalid name to fail")
	}
	if !strings.Contains(err.Error(), "folder name rule") {
		t.Fatalf("error = %q, want folder name rule", err)
	}
}

func TestRunInitReturnsExistingFolderError(t *testing.T) {
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.Mkdir("site", os.ModePerm); err != nil {
		t.Fatal(err)
	}

	err = runInit("site")
	if err == nil {
		t.Fatal("expected existing folder to fail")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("error = %q, want already exists", err)
	}
}

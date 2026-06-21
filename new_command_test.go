package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNewPostCreatesPostInCatalog(t *testing.T) {
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

	err = runNewPost("hello world.md", "notes")
	if err != nil {
		t.Fatalf("runNewPost returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "posts", "notes", "hello-world.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "title: hello world") {
		t.Fatalf("post content = %q, want title", content)
	}
	if !strings.Contains(content, "categories: notes") {
		t.Fatalf("post content = %q, want catalog", content)
	}
}

func TestRunNewPageCreatesPage(t *testing.T) {
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

	err = runNewPage("about")
	if err != nil {
		t.Fatalf("runNewPage returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "pages", "about.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "title: about") {
		t.Fatalf("page content = %q, want title", string(data))
	}
}

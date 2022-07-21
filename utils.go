package main

import (
	"fmt"
	"os"
	"os/exec"
)

func cloneTemplate(base string) error {
	_, err := exec.Command("git", "clone", "https://gitter.top/mder/template", base).Output()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "get latest template failed: %v\n", err)
		return err
	}
	return nil
}

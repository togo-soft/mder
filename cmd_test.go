package main

import (
	"context"
	"errors"
	"flag"
	"testing"
)

func TestNewAppRegistersCommands(t *testing.T) {
	t.Parallel()

	app := newApp()
	want := map[string][]string{
		"init":     nil,
		"generate": []string{"g"},
		"new":      nil,
		"serve":    []string{"s"},
		"update":   []string{"u"},
	}

	got := make(map[string][]string)
	for _, command := range app.Commands {
		got[command.Name] = command.Aliases
	}

	for name, aliases := range want {
		if _, ok := got[name]; !ok {
			t.Fatalf("command %q not registered; got %v", name, got)
		}
		if len(aliases) != len(got[name]) {
			t.Fatalf("command %q aliases = %v, want %v", name, got[name], aliases)
		}
		for i := range aliases {
			if got[name][i] != aliases[i] {
				t.Fatalf("command %q aliases = %v, want %v", name, got[name], aliases)
			}
		}
	}
}

func TestGenerateAliasParsesPathFlag(t *testing.T) {
	tmp := t.TempDir()
	oldBaseDir := BaseDir
	t.Cleanup(func() { BaseDir = oldBaseDir })

	err := newApp().Run(context.Background(), []string{"mder", "g", "--path", tmp})
	if err == nil {
		t.Fatal("expected generate to fail on missing project files")
	}
	if BaseDir != tmp {
		t.Fatalf("BaseDir = %q, want %q", BaseDir, tmp)
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "success", err: nil, want: exitOK},
		{name: "usage", err: flag.ErrHelp, want: exitUsage},
		{name: "runtime", err: errors.New("failed"), want: exitFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := exitCode(tt.err); got != tt.want {
				t.Fatalf("exitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

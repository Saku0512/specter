package main

import (
	"strings"
	"testing"
)

func TestVersionOutput(t *testing.T) {
	orig := version
	version = "1.2.3"
	t.Cleanup(func() {
		version = orig
	})

	got := versionOutput()

	if got == "" {
		t.Fatal("expected non-empty version output")
	}
	if want := "👻 specter 1.2.3"; !strings.HasPrefix(got, want) {
		t.Fatalf("expected output to start with %q, got %q", want, got)
	}
	if want := "____  ____  _____ ____"; !strings.Contains(got, want) {
		t.Fatalf("expected ASCII art to contain %q, got %q", want, got)
	}
}

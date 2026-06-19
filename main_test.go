package main

import (
	"io"
	"os"
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

func TestUsageMentionsStoreFile(t *testing.T) {
	var b strings.Builder
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = orig
	})

	usage()
	w.Close()
	if _, err := io.Copy(&b, r); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	for _, want := range []string{"--store-file", "SPECTER_STORE_FILE"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected usage to mention %q, got %q", want, got)
		}
	}
}

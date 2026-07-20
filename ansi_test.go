package main

import (
	"strings"
	"testing"
)

func TestHasANSI(t *testing.T) {
	if !hasANSI("\x1b[1;35mhi\x1b[0m") {
		t.Fatal("expected ANSI")
	}
	if hasANSI("plain text") {
		t.Fatal("did not expect ANSI")
	}
}

func TestAnsiToHTML(t *testing.T) {
	in := "\x1b[1;35mHEADER\x1b[0m\n\x1b[1;33m[WARN]\x1b[0m x\n\x1b[1;31m[CRITICAL]\x1b[0m y"
	out := ansiToHTML(in)
	wantParts := []string{
		`font-weight:700;color:#d946ef`,
		`HEADER`,
		`color:#eab308`,
		`[WARN]`,
		`color:#ef4444`,
		`[CRITICAL]`,
	}
	for _, w := range wantParts {
		if !strings.Contains(out, w) {
			t.Fatalf("missing %q in %s", w, out)
		}
	}
	if strings.Contains(out, "\x1b") {
		t.Fatal("escape byte leaked into HTML")
	}
}

func TestAnsiToHTMLEscapes(t *testing.T) {
	out := ansiToHTML("<script>alert(1)</script>")
	if strings.Contains(out, "<script>") {
		t.Fatal("HTML not escaped")
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Fatal("expected escaped script tag")
	}
}

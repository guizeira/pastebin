package main

import "testing"

func TestNormalizeLang(t *testing.T) {
	cases := map[string]string{
		"bash":    "bash",
		"SH":      "bash",
		"python":  "python",
		"py":      "python",
		"sql":     "sql",
		"mysql":   "sql",
		"unknown": "bash",
		"":        "bash",
	}
	for in, want := range cases {
		if got := normalizeLang(in); got != want {
			t.Fatalf("normalizeLang(%q)=%q want %q", in, got, want)
		}
	}
}

func TestResolveLang(t *testing.T) {
	if got := resolveLang("python", "bash", "print(1)"); got != "python" {
		t.Fatalf("query should win, got %q", got)
	}
	if got := resolveLang("", "sql", "select 1"); got != "sql" {
		t.Fatalf("stored should win, got %q", got)
	}
	if got := resolveLang("", "", "\x1b[31mred\x1b[0m"); got != "bash" {
		t.Fatalf("ANSI content should default bash, got %q", got)
	}
}

func TestPasteURL(t *testing.T) {
	want := "https://paste.gfonseca.online/abc123?lang=python&fullscreen=1"
	if got := pasteURL("abc123", "py"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

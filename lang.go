package main

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultLang = "bash"

var supportedLangs = map[string]bool{
	"bash":   true,
	"sql":    true,
	"python": true,
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch lang {
	case "py", "python3":
		lang = "python"
	case "sh", "shell", "zsh":
		lang = "bash"
	case "pgsql", "mysql", "mariadb":
		lang = "sql"
	}
	if supportedLangs[lang] {
		return lang
	}
	return defaultLang
}

func langFilePath(id string) string {
	return filepath.Join(dataDir, id+".lang")
}

func savePasteLang(id, lang string) error {
	lang = normalizeLang(lang)
	return os.WriteFile(langFilePath(id), []byte(lang+"\n"), 0644)
}

func loadPasteLang(id string) string {
	b, err := os.ReadFile(langFilePath(id))
	if err != nil {
		return ""
	}
	return normalizeLang(string(b))
}

func resolveLang(queryLang, storedLang, content string) string {
	if q := strings.TrimSpace(queryLang); q != "" {
		return normalizeLang(q)
	}
	if storedLang != "" {
		return normalizeLang(storedLang)
	}
	if hasANSI(content) {
		return "bash"
	}
	return defaultLang
}

func pasteURL(id, lang string) string {
	lang = normalizeLang(lang)
	return "https://paste.gfonseca.online/" + id + "?lang=" + lang + "&fullscreen=1"
}

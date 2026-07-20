package main

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	dataDir       = "./data"
	staticDir     = "./static"
	tplDir        = "./templates"
	assetVersion  = "20260720e"
)

var errorMessages = map[string]map[int]string{
	"en": {
		http.StatusInternalServerError: "Internal server error",
		http.StatusNotFound:            "Paste not found or expired",
	},
	"pt": {
		http.StatusInternalServerError: "Erro interno do servidor",
		http.StatusNotFound:            "Paste não encontrado ou expirado",
	},
}

var tplFuncs = template.FuncMap{
	"asset": func(path string) string { return path + "?v=" + assetVersion },
}

const (
	tplIndexName = "index.html"
	tplViewName  = "view.html"
)

var tplIndex = template.Must(
	template.New(tplIndexName).Funcs(tplFuncs).ParseFiles(filepath.Join(tplDir, tplIndexName)),
)
var tplView = template.Must(
	template.New(tplViewName).Funcs(tplFuncs).ParseFiles(filepath.Join(tplDir, tplViewName)),
)

func renderHTML(w http.ResponseWriter, r *http.Request, tpl *template.Template, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		println("template render error [" + name + "]:", err.Error())
		httpError(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func localeFromRequest(r *http.Request) string {
	accept := strings.ToLower(r.Header.Get("Accept-Language"))
	if strings.HasPrefix(accept, "pt") {
		return "pt"
	}
	return "en"
}

func httpError(w http.ResponseWriter, r *http.Request, status int, fallback string) {
	locale := localeFromRequest(r)
	msg := fallback
	if localized, ok := errorMessages[locale][status]; ok {
		msg = localized
	}
	http.Error(w, msg, status)
}

// ---------------- CREATE PASTE (POST) ----------------
func pasteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // Limite de 10MB

	lang := normalizeLang(r.URL.Query().Get("lang"))
	if hdr := r.Header.Get("X-Paste-Lang"); hdr != "" {
		lang = normalizeLang(hdr)
	}

	id := randomID()
	file := filepath.Join(dataDir, id+".txt")

	f, err := os.Create(file)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	_ = savePasteLang(id, lang)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pasteURL(id, lang)))
}

// ---------------- GET PASTE / VIEW ----------------
func getPaste(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	
	// Se for a raiz, mostra o index (criador de pastes)
	if path == "" {
		renderHTML(w, r, tplIndex, tplIndexName, nil)
		return
	}

	// Detecta se o usuário quer a versão RAW pura
	// Aceita tanto "/id/raw" quanto comandos de terminal (cURL/Wget)
	isRaw := strings.HasSuffix(path, "/raw")
	userAgent := strings.ToLower(r.UserAgent())
	isTerminal := strings.Contains(userAgent, "curl") || strings.Contains(userAgent, "wget")

	id := strings.TrimSuffix(path, "/raw")
	file := filepath.Join(dataDir, id+".txt")

	if _, err := os.Stat(file); os.IsNotExist(err) {
		httpError(w, r, http.StatusNotFound, "Paste not found or expired")
		return
	}

	// Se for terminal ou rota /raw, entrega o texto puro original
	if isRaw || isTerminal {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeFile(w, r, file)
		return
	}

	// Caso contrário, lê o conteúdo e injeta no template de visualização premium
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "Internal server error")
		return
	}

	content := string(contentBytes)
	lang := resolveLang(r.URL.Query().Get("lang"), loadPasteLang(id), content)

	view := struct {
		ID          string
		Content     string
		ContentHTML template.HTML
		HasANSI     bool
		Lang        string
		Langs       []string
	}{
		ID:      id,
		Content: content,
		Lang:    lang,
		Langs:   []string{"bash", "sql", "python"},
	}

	// Colored terminal output is independent of the syntax highlighter choice.
	if hasANSI(content) {
		view.HasANSI = true
		view.ContentHTML = template.HTML(ansiToHTML(content))
	}

	renderHTML(w, r, tplView, tplViewName, view)
}

// ---------------- STATIC ----------------
func staticHandler() http.Handler {
	fs := http.FileServer(http.Dir(staticDir))
	return http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		fs.ServeHTTP(w, r)
	}))
}

func main() {
	_ = os.MkdirAll(dataDir, 0755)

	http.HandleFunc("/", getPaste)
	http.HandleFunc("/paste", pasteHandler)
	http.Handle("/static/", staticHandler())

	println("Servidor escutando na porta :8080")
	_ = http.ListenAndServe(":8080", nil)
}

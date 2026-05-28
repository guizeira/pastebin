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
	dataDir   = "./data"
	staticDir = "./static"
	tplDir    = "./templates"
)

// Carrega os dois templates necessários
var tplIndex = template.Must(template.ParseFiles(filepath.Join(tplDir, "index.html")))
var tplView = template.Must(template.ParseFiles(filepath.Join(tplDir, "view.html")))

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------------- CREATE PASTE (POST) ----------------
func pasteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // Limite de 10MB

	id := randomID()
	file := filepath.Join(dataDir, id+".txt")
	
	f, err := os.Create(file)
	if err != nil {
		http.Error(w, "Erro interno ao criar o registro", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, "Erro ao gravar os dados enviados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("https://paste.gfonseca.online/" + id))
}

// ---------------- GET PASTE / VIEW ----------------
func getPaste(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	
	// Se for a raiz, mostra o index (criador de pastes)
	if path == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tplIndex.Execute(w, nil)
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
		http.Error(w, "Paste não encontrado ou expirado", http.StatusNotFound)
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
		http.Error(w, "Erro ao ler o conteúdo", http.StatusInternalServerError)
		return
	}

	data := map[string]string{
		"ID":      id,
		"Content": string(contentBytes),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tplView.Execute(w, data)
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

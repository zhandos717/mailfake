package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"fake-mail/internal/store"
)

//go:embed templates/*
var templatesFS embed.FS

// Server handles HTTP requests
type Server struct {
	addr  string
	store *store.Store
}

// New creates a new web server
func New(addr string, emailStore *store.Store) *Server {
	return &Server{
		addr:  addr,
		store: emailStore,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/email/", s.emailHandler)
	http.HandleFunc("/html/", s.htmlHandler)
	http.HandleFunc("/api/emails", s.apiEmailsHandler)
	http.HandleFunc("/api/emails/", s.apiDeleteHandler)
	http.HandleFunc("/api/clear", s.apiClearHandler)

	log.Printf("Web interface available at http://localhost%s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := r.URL.Query().Get("q")
	emails := s.store.Search(query)

	tmpl.Execute(w, map[string]interface{}{
		"Emails": emails,
		"Count":  len(emails),
		"Total":  s.store.Count(),
		"Query":  query,
	})
}

func (s *Server) emailHandler(w http.ResponseWriter, r *http.Request) {
	var id int
	if _, err := fmt.Sscanf(r.URL.Path, "/email/%d", &id); err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	email := s.store.Get(id)
	if email == nil {
		http.Error(w, "Email not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFS(templatesFS, "templates/email.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, email)
}

func (s *Server) htmlHandler(w http.ResponseWriter, r *http.Request) {
	var id int
	if _, err := fmt.Sscanf(r.URL.Path, "/html/%d", &id); err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	email := s.store.Get(id)
	if email == nil {
		http.Error(w, "Email not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if email.HTMLBody != "" {
		w.Write([]byte(email.HTMLBody))
	} else {
		// Обернём текст в базовый HTML
		w.Write([]byte("<html><body><pre style=\"font-family: sans-serif; white-space: pre-wrap;\">" + email.TextBody + "</pre></body></html>"))
	}
}

func (s *Server) apiEmailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.store.GetAll())
}

func (s *Server) apiDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	if _, err := fmt.Sscanf(r.URL.Path, "/api/emails/%d", &id); err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	if s.store.Delete(id) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		http.Error(w, "Email not found", http.StatusNotFound)
	}
}

func (s *Server) apiClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.store.Clear()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

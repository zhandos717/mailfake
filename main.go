package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
)

//go:embed templates/*
var templatesFS embed.FS

// Email represents a received email
type Email struct {
	ID        int       `json:"id"`
	From      string    `json:"from"`
	To        []string  `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	RawData   string    `json:"raw_data"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailStore stores all received emails
type EmailStore struct {
	mu     sync.RWMutex
	emails []Email
	nextID int
}

var store = &EmailStore{
	emails: make([]Email, 0),
	nextID: 1,
}

func (s *EmailStore) Add(email Email) {
	s.mu.Lock()
	defer s.mu.Unlock()
	email.ID = s.nextID
	email.CreatedAt = time.Now()
	s.nextID++
	s.emails = append([]Email{email}, s.emails...)
}

func (s *EmailStore) GetAll() []Email {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Email, len(s.emails))
	copy(result, s.emails)
	return result
}

func (s *EmailStore) Get(id int) *Email {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, e := range s.emails {
		if e.ID == id {
			emailCopy := e
			return &emailCopy
		}
	}
	return nil
}

func (s *EmailStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emails = make([]Email, 0)
}

func (s *EmailStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.emails {
		if e.ID == id {
			s.emails = append(s.emails[:i], s.emails[i+1:]...)
			return true
		}
	}
	return false
}

// SMTP Backend
type Backend struct{}

func (bkd *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{}, nil
}

// Session handles SMTP session
type Session struct {
	from string
	to   []string
}

func (s *Session) AuthPlain(username, password string) error {
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	rawData := string(data)
	subject, body := parseEmail(rawData)

	email := Email{
		From:    s.from,
		To:      s.to,
		Subject: subject,
		Body:    body,
		RawData: rawData,
	}

	store.Add(email)
	log.Printf("Received email from %s to %v: %s", s.from, s.to, subject)
	return nil
}

func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *Session) Logout() error {
	return nil
}

func parseEmail(rawData string) (subject, body string) {
	lines := strings.Split(strings.ReplaceAll(rawData, "\r\n", "\n"), "\n")
	inBody := false
	var bodyLines []string

	for _, line := range lines {
		if inBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		if line == "" {
			inBody = true
			continue
		}

		if strings.HasPrefix(strings.ToLower(line), "subject:") {
			subject = strings.TrimSpace(line[8:])
		}
	}

	body = strings.Join(bodyLines, "\n")
	return
}

// HTTP Handlers
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	emails := store.GetAll()
	tmpl.Execute(w, map[string]interface{}{
		"Emails": emails,
		"Count":  len(emails),
	})
}

func emailHandler(w http.ResponseWriter, r *http.Request) {
	var id int
	_, err := fmt.Sscanf(r.URL.Path, "/email/%d", &id)
	if err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	email := store.Get(id)
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

func apiEmailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.GetAll())
}

func apiClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	store.Clear()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func apiDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var id int
	_, err := fmt.Sscanf(r.URL.Path, "/api/emails/%d", &id)
	if err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	if store.Delete(id) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		http.Error(w, "Email not found", http.StatusNotFound)
	}
}

func main() {
	smtpPort := "1025"
	httpPort := "8025"

	// Start SMTP server
	be := &Backend{}
	s := smtp.NewServer(be)
	s.Addr = ":" + smtpPort
	s.Domain = "localhost"
	s.AllowInsecureAuth = true

	go func() {
		log.Printf("SMTP server listening on :%s", smtpPort)
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Start HTTP server
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/email/", emailHandler)
	http.HandleFunc("/api/emails", apiEmailsHandler)
	http.HandleFunc("/api/clear", apiClearHandler)
	http.HandleFunc("/api/emails/", apiDeleteHandler)

	log.Printf("Web interface available at http://localhost:%s", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}

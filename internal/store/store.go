package store

import (
	"strings"
	"sync"
	"time"
)

// Email represents a received email
type Email struct {
	ID          int       `json:"id"`
	From        string    `json:"from"`
	To          []string  `json:"to"`
	Subject     string    `json:"subject"`
	TextBody    string    `json:"text_body"`
	HTMLBody    string    `json:"html_body"`
	ContentType string    `json:"content_type"`
	RawData     string    `json:"raw_data"`
	CreatedAt   time.Time `json:"created_at"`
}

// HasHTML returns true if email has HTML content
func (e *Email) HasHTML() bool {
	return e.HTMLBody != ""
}

// Store manages email storage
type Store struct {
	mu     sync.RWMutex
	emails []Email
	nextID int
}

// New creates a new email store
func New() *Store {
	return &Store{
		emails: make([]Email, 0),
		nextID: 1,
	}
}

// Add adds a new email to the store
func (s *Store) Add(email Email) {
	s.mu.Lock()
	defer s.mu.Unlock()
	email.ID = s.nextID
	email.CreatedAt = time.Now()
	s.nextID++
	s.emails = append([]Email{email}, s.emails...)
}

// GetAll returns all emails
func (s *Store) GetAll() []Email {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Email, len(s.emails))
	copy(result, s.emails)
	return result
}

// Get returns an email by ID
func (s *Store) Get(id int) *Email {
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

// Delete removes an email by ID
func (s *Store) Delete(id int) bool {
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

// Clear removes all emails
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.emails = make([]Email, 0)
}

// Count returns the number of emails
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.emails)
}

// Search filters emails by query (matches From, To, Subject)
func (s *Store) Search(query string) []Email {
	if query == "" {
		return s.GetAll()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	var result []Email

	for _, e := range s.emails {
		if strings.Contains(strings.ToLower(e.From), query) ||
			strings.Contains(strings.ToLower(e.Subject), query) ||
			containsAny(e.To, query) {
			result = append(result, e)
		}
	}

	return result
}

func containsAny(slice []string, query string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), query) {
			return true
		}
	}
	return false
}

package smtp

import (
	"encoding/base64"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/zhandos717/mailfake/internal/store"

	"github.com/emersion/go-smtp"
)

// Server wraps the SMTP server
type Server struct {
	server *smtp.Server
	store  *store.Store
}

// New creates a new SMTP server
func New(addr string, emailStore *store.Store) *Server {
	s := &Server{store: emailStore}

	backend := &backend{store: emailStore}
	s.server = smtp.NewServer(backend)
	s.server.Addr = addr
	s.server.Domain = "localhost"
	s.server.AllowInsecureAuth = true

	return s
}

// Start starts the SMTP server
func (s *Server) Start() error {
	log.Printf("SMTP server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// backend implements smtp.Backend
type backend struct {
	store *store.Store
}

func (b *backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &session{store: b.store}, nil
}

// session handles a single SMTP session
type session struct {
	store *store.Store
	from  string
	to    []string
}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	rawData := string(data)
	parsed := parseEmail(rawData)

	email := store.Email{
		From:        s.from,
		To:          s.to,
		Subject:     parsed.Subject,
		TextBody:    parsed.TextBody,
		HTMLBody:    parsed.HTMLBody,
		ContentType: parsed.ContentType,
		RawData:     rawData,
	}

	s.store.Add(email)
	log.Printf("Received email from %s to %v: %s", s.from, s.to, parsed.Subject)
	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}

type parsedEmail struct {
	Subject     string
	TextBody    string
	HTMLBody    string
	ContentType string
}

func parseEmail(rawData string) parsedEmail {
	result := parsedEmail{}

	msg, err := mail.ReadMessage(strings.NewReader(rawData))
	if err != nil {
		// Fallback: простой парсинг
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
				result.Subject = strings.TrimSpace(line[8:])
			}
		}
		result.TextBody = strings.Join(bodyLines, "\n")
		return result
	}

	// Subject
	result.Subject = decodeHeader(msg.Header.Get("Subject"))

	// Content-Type
	contentType := msg.Header.Get("Content-Type")
	result.ContentType = contentType

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Простое текстовое письмо
		body, _ := io.ReadAll(msg.Body)
		result.TextBody = decodeBody(string(body), msg.Header.Get("Content-Transfer-Encoding"))
		return result
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		// MIME multipart
		parseMultipart(msg.Body, params["boundary"], &result)
	} else if strings.HasPrefix(mediaType, "text/html") {
		body, _ := io.ReadAll(msg.Body)
		result.HTMLBody = decodeBody(string(body), msg.Header.Get("Content-Transfer-Encoding"))
	} else {
		body, _ := io.ReadAll(msg.Body)
		result.TextBody = decodeBody(string(body), msg.Header.Get("Content-Transfer-Encoding"))
	}

	return result
}

func parseMultipart(r io.Reader, boundary string, result *parsedEmail) {
	mr := multipart.NewReader(r, boundary)

	for {
		part, err := mr.NextPart()
		if err != nil {
			break
		}

		contentType := part.Header.Get("Content-Type")
		encoding := part.Header.Get("Content-Transfer-Encoding")
		body, _ := io.ReadAll(part)
		decoded := decodeBody(string(body), encoding)

		mediaType, params, _ := mime.ParseMediaType(contentType)

		if strings.HasPrefix(mediaType, "multipart/") {
			// Вложенный multipart
			parseMultipart(strings.NewReader(decoded), params["boundary"], result)
		} else if strings.HasPrefix(mediaType, "text/html") {
			result.HTMLBody = decoded
		} else if strings.HasPrefix(mediaType, "text/plain") {
			result.TextBody = decoded
		}
	}
}

func decodeBody(body, encoding string) string {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	switch encoding {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(body, "\n", ""))
		if err != nil {
			return body
		}
		return string(decoded)
	case "quoted-printable":
		decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
		if err != nil {
			return body
		}
		return string(decoded)
	default:
		return body
	}
}

func decodeHeader(header string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(header)
	if err != nil {
		return header
	}
	return decoded
}

# Fake Mail

Локальный SMTP сервер для тестирования отправки email с веб-интерфейсом. Альтернатива Mailhog/Mailtrap.

## Возможности

- SMTP сервер без аутентификации
- Веб-интерфейс для просмотра писем
- Поддержка HTML и текстовых писем
- Поиск по адресам и теме
- REST API
- Копирование конфига одним кликом

## Запуск

```bash
# Через go run
go run ./cmd

# Или через make
make run

# Или собрать и запустить
make build
./bin/fake-mail
```

После запуска:
- **SMTP**: `localhost:1025`
- **Web**: http://localhost:8025

## Настройка

### Laravel

```env
MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
```

### Django

```python
# settings.py
EMAIL_BACKEND = 'django.core.mail.backends.smtp.EmailBackend'
EMAIL_HOST = '127.0.0.1'
EMAIL_PORT = 1025
EMAIL_USE_TLS = False
```

### Node.js (Nodemailer)

```javascript
const transporter = nodemailer.createTransport({
  host: '127.0.0.1',
  port: 1025,
  secure: false,
});
```

### Ruby on Rails

```ruby
# config/environments/development.rb
config.action_mailer.delivery_method = :smtp
config.action_mailer.smtp_settings = {
  address: '127.0.0.1',
  port: 1025
}
```

### Go

```go
import "net/smtp"

smtp.SendMail("127.0.0.1:1025", nil, "from@test.com",
    []string{"to@test.com"}, []byte(message))
```

### Python

```python
import smtplib

with smtplib.SMTP('127.0.0.1', 1025) as server:
    server.sendmail('from@test.com', 'to@test.com', message)
```

### Symfony

```yaml
# config/packages/mailer.yaml
framework:
  mailer:
    dsn: 'smtp://127.0.0.1:1025'
```

### Spring Boot

```properties
# application.properties
spring.mail.host=127.0.0.1
spring.mail.port=1025
```

## API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/` | Веб-интерфейс |
| GET | `/?q=search` | Поиск писем |
| GET | `/api/emails` | Список писем (JSON) |
| GET | `/html/{id}` | HTML контент письма |
| DELETE | `/api/emails/{id}` | Удалить письмо |
| POST | `/api/clear` | Очистить все |

## Структура проекта

```
fake-mail/
├── cmd/
│   └── main.go
├── internal/
│   ├── smtp/
│   │   └── server.go
│   ├── store/
│   │   └── store.go
│   └── web/
│       ├── server.go
│       └── templates/
├── Makefile
└── README.md
```

## Make команды

```bash
make build       # Сборка
make run         # Запуск
make clean       # Очистка
make build-all   # Сборка для Linux/Mac/Windows
```

## Docker

```bash
# Docker Compose
docker-compose up -d

# Или вручную
docker build -t fake-mail .
docker run -p 1025:1025 -p 8025:8025 fake-mail
```

## Лицензия

MIT

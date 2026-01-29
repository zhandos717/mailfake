# Fake Mail

Простой SMTP сервер для тестирования отправки email из Laravel приложений с веб-интерфейсом.

## Возможности

- SMTP сервер на порту `1025`
- Веб-интерфейс на порту `8025`
- Просмотр всех полученных писем
- Просмотр raw данных письма
- Удаление отдельных писем и очистка всех
- REST API для интеграции
- Автообновление списка каждые 5 секунд

## Установка и запуск

```bash
# Клонировать или перейти в директорию
cd fake-mail

# Установить зависимости
go mod tidy

# Запустить
go run main.go
```

После запуска:
- SMTP сервер: `localhost:1025`
- Веб-интерфейс: http://localhost:8025

## Настройка Laravel

В файле `.env` вашего Laravel приложения:

```env
MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS="test@example.com"
MAIL_FROM_NAME="${APP_NAME}"
```

## Пример отправки email в Laravel

```php
use Illuminate\Support\Facades\Mail;

// Простая отправка
Mail::raw('Текст письма', function ($message) {
    $message->to('user@example.com')
            ->subject('Тестовое письмо');
});

// Через Mailable
Mail::to('user@example.com')->send(new WelcomeMail());
```

Или через Artisan Tinker:

```bash
php artisan tinker
>>> Mail::raw('Test', fn($m) => $m->to('test@test.com')->subject('Hello'));
```

## API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/emails` | Получить все письма (JSON) |
| DELETE | `/api/emails/{id}` | Удалить письмо |
| POST | `/api/clear` | Удалить все письма |

### Примеры использования API

```bash
# Получить все письма
curl http://localhost:8025/api/emails

# Удалить письмо с ID 1
curl -X DELETE http://localhost:8025/api/emails/1

# Очистить все письма
curl -X POST http://localhost:8025/api/clear
```

## Сборка бинарника

```bash
go build -o fake-mail .
./fake-mail
```

## Docker (опционально)

Создайте `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o fake-mail .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/fake-mail .
COPY --from=builder /app/templates ./templates
EXPOSE 1025 8025
CMD ["./fake-mail"]
```

Запуск:

```bash
docker build -t fake-mail .
docker run -p 1025:1025 -p 8025:8025 fake-mail
```

## Лицензия

MIT

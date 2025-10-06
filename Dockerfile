# Базовый образ для сборки
FROM golang:1.21-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git ca-certificates

# Создание рабочей директории
WORKDIR /app

# Копирование файлов модулей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o tgBotRecommender

# Финальный образ
FROM alpine:latest

# Установка зависимостей для работы
RUN apk --no-cache add ca-certificates tzdata

# Создание non-root пользователя для безопасности
RUN adduser -D -g '' appuser

# Копирование бинарного файла из builder
COPY --from=builder /app/tgBotRecommender /app/tgBotRecommender

# Копирование SQL файлов
COPY --from=builder /app/storage/database/init.sql /app/init.sql

# Установка прав
RUN chown -R appuser:appuser /app

# Переключение на non-root пользователя
USER appuser

# Установка рабочей директории
WORKDIR /app

# Экспортируем порт для Render
EXPOSE 8080

# Команда для запуска приложения
CMD ["
/app/tgBotRecommender"]
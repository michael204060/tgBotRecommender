# Базовый образ для сборки
FROM golang:1.24-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git

# Создание рабочей директории
WORKDIR /app

# Копирование файлов модулей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/tgBotRecommender

# Финальный образ
FROM alpine:latest

# Установка зависимостей для работы с PostgreSQL
RUN apk add --no-cache libc6-compat

# Копирование бинарного файла из builder
COPY --from=builder /app/tgBotRecommender /app/tgBotRecommender

# Установка рабочей директории
WORKDIR /app

# Команда для запуска приложения
CMD ["/app/tgBotRecommender"]
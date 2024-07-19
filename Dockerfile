# Используем официальный образ Golang
FROM golang:1.22

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum файлы и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные файлы проекта
COPY . .

# Сборка приложения
RUN go build -o tgBotRecommender ./cmd/tgBotRecommender

# Устанавливаем переменную окружения для токена
ENV TG_BOT_TOKEN='7047428650:AAGnJCnA_RUZJ0TFntTYKqVYApD0vuQKNls'

# Указываем команду для запуска приложения
CMD ["./tgBotRecommender", "-tg-bot-token", "$TG_BOT_TOKEN"]

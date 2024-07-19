# Используем официальный образ Golang
FROM golang:1.22

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum файлы и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем все файлы проекта в рабочую директорию контейнера
COPY . .

# Сборка приложения
RUN go build -o tgBotRecommender main.go

# Указываем команду для запуска приложения
CMD ["./tgBotRecommender", "-tg-bot-token", "$TG_BOT_TOKEN"]

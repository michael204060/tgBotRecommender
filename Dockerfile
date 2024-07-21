# Используем официальный образ Golang
FROM golang:1.22

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и скачиваем зависимости
COPY go.mod ./
RUN go mod download

# Копируем все файлы проекта в рабочую директорию контейнера
COPY . .

# Сборка приложения
RUN go build -o tgBotRecommender main.go

# Устанавливаем переменную окружения для порта
ENV PORT=8080
EXPOSE $PORT

# Указываем команду для запуска приложения
CMD ["./tgBotRecommender", "-tg-bot-token", "7047428650:AAGnJCnA_RUZJ0TFntTYKqVYApD0vuQKNls"]

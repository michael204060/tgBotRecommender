# Используем официальный образ Golang
FROM golang:1.22

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и скачиваем зависимости
COPY go.mod ./
RUN go mod tidy

# Копируем все файлы проекта в рабочую директорию контейнера
COPY . .

# Изменение прав доступа и создание папки files_storage
RUN chmod +x /bin/mkdir && mkdir -p /app/files_storage

# Сборка приложения
RUN go build -o tgBotRecommender main.go

# Устанавливаем переменную окружения для порта
ENV PORT=8080
EXPOSE $PORT

# Указываем команду для запуска приложения
CMD ["./tgBotRecommender", "-tg-bot-token", "YOUR_BOT_TOKEN_HERE"]

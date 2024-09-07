# Используем официальный образ Golang
FROM golang:1.23.0

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum и скачиваем зависимости
COPY go.mod ./
RUN go mod tidy

# Копируем все файлы проекта в рабочую директорию контейнера
COPY . .

# Создаем папку files_storage, изменяем права доступа и выполняем сборку приложения
RUN mkdir -p /app/files_storage && \
    chmod -R 777 /app/files_storage && \
    go build -o tgBotRecommender main.go

# Устанавливаем переменную окружения для порта
ENV PORT=8080
EXPOSE $PORT

# Указываем команду для запуска приложения
CMD ["./tgBotRecommender", "-tg-bot-token", "YOUR_BOT_TOKEN_HERE"]

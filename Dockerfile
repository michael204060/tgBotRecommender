FROM golang:1.23.0
WORKDIR /app
COPY go.mod ./
RUN go mod tidy
COPY . .
RUN mkdir -p /app/files_storage && \
    chmod -R 777 /app/files_storage && \
    go build -o tgBotRecommender main.go
ENV PORT=8080
EXPOSE $PORT
CMD ["./tgBotRecommender", "-tg-bot-token", "7047428650:AAGnJCnA_RUZJ0TFntTYKqVYApD0vuQKNls"]

FROM golang:1.25.6-alpine

WORKDIR /app

# Копируем go.mod и go.sum из корня проекта
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем бинарник
RUN go build -o bot ./cmd/bot

CMD ["./bot"]

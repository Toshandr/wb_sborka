# Этап 1: Сборка приложения
FROM golang:1.26.0 AS builder

WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и собираем бинарный файл
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/wms-api .

# Этап 2: Финальный минимальный образ
FROM alpine:latest

WORKDIR /app

# Копируем собранный бинарный файл из первого этапа
COPY --from=builder /app/wms-api .

# Открываем порт, который слушает приложение
EXPOSE 8080

# Команда для запуска приложения
ENTRYPOINT ["/app/wms-api"]
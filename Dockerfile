FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .
RUN swag init -g ./cmd/api/main.go -o ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

FROM scratch

COPY --from=builder /app/app /app
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["/app"]

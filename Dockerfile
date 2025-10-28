# Build stage
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application from the correct path
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/api

# Final stage
FROM alpine:latest

# Install curl for healthcheck and ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/app .

# Copy docs if they exist
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./app"]

# Backend Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod first (go.sum may not exist yet)
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Ensure module files are up to date and build the application
RUN go mod tidy \
 && CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]

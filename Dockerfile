FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy all source
COPY . .

# Download dependencies and build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /api .

EXPOSE 8080

CMD ["./api"]

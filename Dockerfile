FROM golang:1.25.5-alpine AS builder
WORKDIR /backend
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o backend main.go

# Runtime stage
FROM alpine:latest
RUN apk update && apk upgrade --no-cache libexpat
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /backend/backend .
EXPOSE 9001
CMD ["./backend"]

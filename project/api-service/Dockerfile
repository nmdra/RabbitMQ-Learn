# Use Golang base image
FROM golang:alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

FROM alpine:latest AS production

COPY --from=builder /app/main ./

CMD ["./main"]
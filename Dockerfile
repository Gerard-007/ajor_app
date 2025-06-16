FROM golang:1.24-alpine

# Install dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Install air
RUN go install github.com/air-verse/air@latest

# Copy source code and configuration
COPY . .

# Expose port
EXPOSE 8080

# Run air with .air.toml
CMD ["air", "-c", ".air.toml"]
# CMD ["./main"]  # For production only.

# Development Dockerfile
FROM golang:1.25-alpine

# Install git and essential tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Install Air for Hot Reload
RUN go install github.com/air-verse/air@latest

# Copy dependency files first (Caching layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Run Air (Live Reload)
CMD ["air", "-c", ".air.toml"]
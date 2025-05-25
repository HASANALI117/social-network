# Use a Golang base image
FROM golang:1.24-alpine3.21

# Set the working directory
WORKDIR /app

# Install dependencies for CGo (SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the backend source code
COPY . .

# Build with verbose output to see any errors
RUN go build -v -o main ./cmd/server/main.go

# Ensure the binary is executable
RUN chmod +x main

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]
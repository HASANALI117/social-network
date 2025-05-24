# Use a Golang base image
FROM golang:1.23.8-alpine3.20

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

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["go", "run", "./cmd/server/main.go"]
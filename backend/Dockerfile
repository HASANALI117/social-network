FROM golang:1.23.6
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o social-network ./cmd/server/main.go
CMD ["./social-network"]
EXPOSE 8080
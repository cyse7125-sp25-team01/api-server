# Use the official Golang image as the base image
FROM golang:latest AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules and Download Dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the Source Code into the Container
COPY . .

# Build the Go app (statically compiled for compatibility with Alpine)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api ./cmd/api

# Start a new stage to create the final image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/api .

# Expose the port the app will run on
EXPOSE 8080

# Command to run the application
CMD ["./api"]

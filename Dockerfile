# Build stage
FROM golang:1.22.3-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Create an empty log.txt file
RUN touch log.txt

# Final stage
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built application from the build stage
COPY --from=build /app/main .
COPY --from=build /app/log.txt .

# Ensure log.txt is writable by the application
RUN chmod 666 log.txt

# Expose the port the app runs on
EXPOSE 8080

# Run the Go application
CMD ["./main"]

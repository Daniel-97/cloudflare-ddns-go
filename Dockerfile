# Stage 1: Build
FROM golang:1.24.5 AS builder

WORKDIR /app

# Copy the source files
COPY . .

# Build the Go binary
RUN go build -o -v app

# Stage 2: Minimal runtime image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/app .

# Run the app
CMD ["app"]

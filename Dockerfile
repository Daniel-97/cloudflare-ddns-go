# Stage 1: Build
FROM golang:1.24.5

WORKDIR /build

# Copy the source files
COPY . .

# Build the Go binary
RUN go build -v -o app

# Run the app
CMD ["/build/app"]

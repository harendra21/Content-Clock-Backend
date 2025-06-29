# ---------- STAGE 1: Build ----------
FROM golang:1.24-alpine AS builder

# Install unzip (needed for PB release extraction if applicable)
RUN apk add --no-cache git build-base

WORKDIR /app

# Copy go.mod and go.sum files to leverage Docker cache
COPY go.mod go.sum ./

# Tidy up modules (assuming your go.mod is set up correctly)
RUN go mod tidy

# Copy your Go source files
COPY  . .

# Build PocketBase binary (static)
RUN go build -o pb .

# ---------- STAGE 2: Runtime ----------
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /pb

# Copy the built binary from builder
COPY --from=builder /app/pb .

# Expose default PocketBase port
EXPOSE 8080

# Run PocketBase
CMD ["./pb", "serve", "--http=0.0.0.0:8080"]
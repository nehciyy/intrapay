# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM --platform=$BUILDPLATFORM golang:1.23.3 AS build
WORKDIR /src

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source
COPY . .

# Build the Go app from /cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/intrapay ./cmd/server

# ---- Final minimal image ----
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
ARG UID=10001
RUN adduser -u ${UID} -D -g '' appuser
USER appuser

# Copy binary from builder stage
COPY --from=build /bin/intrapay /bin/intrapay

# Expose app port
EXPOSE 8080

# Run the server
ENTRYPOINT ["/bin/intrapay"]
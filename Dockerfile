# Build stage using Go 1.25
FROM golang:1.25-bookworm AS build

# Install only essential tools (no SQLite)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/MyTeleProject2026/Slotopol-server

# Copy only go.mod – go.sum will be generated
COPY go.mod ./

# Download dependencies AND tidy to generate a complete go.sum
RUN go mod download && go mod tidy

# Copy the rest of the source (go.sum is now present and will NOT be overwritten)
COPY . .

# Make build scripts executable
RUN chmod +x ./task/*.sh

# Build the application (CGO disabled, no SQLite)
RUN ./task/build-docker.sh

# Final lightweight image
FROM scratch

# Copy binary and configuration files
COPY --from=build /go/bin/app /go/bin/app
COPY --from=build /go/src/github.com/MyTeleProject2026/Slotopol-server/appdata /appdata

EXPOSE 8080

ENTRYPOINT ["/go/bin/app"]
CMD ["-v", "web"]

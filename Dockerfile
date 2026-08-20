# Build stage
FROM golang:1.25-bookworm AS build

# Install C libraries for SQLite (and static linking)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/MyTeleProject2026/Slotopol-server

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Make build scripts executable
RUN chmod +x ./task/*.sh

# Compile with all required tags and static linking
RUN ./task/build-docker.sh

# Deploy stage
FROM scratch

# Copy the binary and configuration
COPY --from=build /go/bin/app /go/bin/app
COPY --from=build /go/src/github.com/MyTeleProject2026/Slotopol-server/appdata /appdata

EXPOSE 8080

ENTRYPOINT ["/go/bin/app"]
CMD ["-v", "web"]

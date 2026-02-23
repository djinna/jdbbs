# Build stage
FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /prodcal ./cmd/srv

# Runtime stage
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates sqlite3 && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -r -m -d /app prodcal
WORKDIR /app
COPY --from=builder /prodcal /app/prodcal
COPY seed_data.json /app/seed_data.json

USER prodcal
EXPOSE 8000
VOLUME ["/app/data"]

ENTRYPOINT ["/app/prodcal"]
CMD ["-listen", ":8000"]

# syntax=docker/dockerfile:1.7

# ---------- Build stage ----------
FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /jdbbs ./cmd/srv

# ---------- Runtime stage ----------
FROM debian:bookworm-slim

ARG TYPST_VERSION=0.13.1
ARG TARGETARCH=x86_64

# Runtime deps:
#   - sqlite3 / ca-certificates: prodcal core
#   - pandoc:                    DOCX <-> Markdown / EPUB conversions
#   - python3 + python-docx + pyyaml: word-template + corrections scripts
#   - libertinus-otf:            body font (book-prod assumes system-installed)
#   - fontconfig:                so typst can find bundled & system fonts
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        sqlite3 \
        pandoc \
        python3 \
        python3-docx \
        python3-yaml \
        fonts-libertinus \
        fontconfig \
    && rm -rf /var/lib/apt/lists/*

# Install Typst (statically-linked binary release).
RUN curl -fsSL "https://github.com/typst/typst/releases/download/v${TYPST_VERSION}/typst-${TARGETARCH}-unknown-linux-musl.tar.xz" \
      | tar -xJ -C /tmp \
    && mv "/tmp/typst-${TARGETARCH}-unknown-linux-musl/typst" /usr/local/bin/typst \
    && chmod +x /usr/local/bin/typst \
    && rm -rf "/tmp/typst-${TARGETARCH}-unknown-linux-musl"

RUN useradd -r -m -d /app jdbbs
WORKDIR /app

# Server binary
COPY --from=builder /jdbbs /app/jdbbs

# Bundled typesetting assets (templates, fonts, scripts, filters)
COPY --chown=jdbbs:jdbbs typesetting/ /app/typesetting/

# Optional reference + sample content (small enough to keep)
COPY --chown=jdbbs:jdbbs corrections/ /app/corrections/

# Make bundled fonts visible to fontconfig as well as typst --font-path
RUN cp -r /app/typesetting/fonts/sourcesans/OTF /usr/share/fonts/truetype/sourcesans \
    && fc-cache -f

USER jdbbs
EXPOSE 8000
VOLUME ["/app/data"]

ENV JDBBS_TYPESETTING_DIR=/app/typesetting

ENTRYPOINT ["/app/jdbbs"]
CMD ["-listen", ":8000"]

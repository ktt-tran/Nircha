FROM golang:1.23-bookworm AS builder

COPY ranker/ /ranker/

WORKDIR /src

COPY searcher/go.mod searcher/go.sum ./
RUN go mod download

COPY searcher/ .

RUN go build -o /out/searcher ./cmd

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
&& rm -rf /var/lib/apt/lists/*

RUN useradd --system --no-create-home --shell /usr/sbin/nologin searcher
WORKDIR /app
COPY --from=builder /out/searcher .
USER searcher

ENV PORT=8080
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f "http://localhost:${PORT}/api/health" || exit 1

ENTRYPOINT ["./searcher"]
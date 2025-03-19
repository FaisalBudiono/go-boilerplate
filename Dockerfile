FROM golang:1.23.3 AS base

FROM base AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x

FROM deps AS builder-api
COPY . .
RUN CGO_ENABLED=0 go build -o api ./cmd/api/main.go

FROM deps AS builder-migrator
COPY . .
RUN CGO_ENABLED=0 go build -o migrator ./cmd/migrator/main.go

FROM alpine:3.10 AS api
USER 1000
WORKDIR /app
RUN mkdir logs
COPY db /app/db
COPY --from=base /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
COPY --from=builder-api /app/api /app/api
COPY --from=builder-migrator /app/migrator /app/migrator
CMD ["/app/api"]

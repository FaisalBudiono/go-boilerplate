FROM golang:1.23.3-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 go build -o migrator ./cmd/migrator/main.go
RUN CGO_ENABLED=0 go build -o api ./cmd/api/main.go

FROM alpine:3.10 AS api
USER 1000
WORKDIR /app
COPY --from=build /app/db /app/db
COPY --from=build /app/migrator /app/migrator
COPY --from=build /app/api /app/api
CMD ["/app/api"]


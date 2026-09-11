FROM golang:1.27 AS base

FROM alpine:3.10 AS api
USER 1000
WORKDIR /app
RUN mkdir logs
COPY --from=base /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
COPY bin/api /app/api
COPY bin/migrator /app/migrator
CMD ["/app/api"]

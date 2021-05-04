FROM golang:1.16.2-buster as golang

WORKDIR /bcloud-build

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY agent/main.go agent/main.go
COPY agent/pkg     agent/pkg
COPY gen           gen
RUN CGO_ENABLED=0 GOOS=linux go build -o /bcloud-agent -mod=readonly agent/main.go


FROM debian:buster-20210208-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=golang /bcloud-agent .
ENTRYPOINT ["./bcloud-agent"]

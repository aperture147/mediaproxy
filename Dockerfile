FROM golang:1.15-buster AS build-env

LABEL maintainer="HaiVQ <haivq@house3d.net>"

WORKDIR /app

COPY go.mod go.mod
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=1.0.0 -w -s" -o server

FROM debian:buster-slim

LABEL maintainer="HaiVQ <haivq@house3d.net>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ENV AWS_ACCESS_KEY_ID=placeholder \
    AWS_SECRET_ACCESS_KEY=placeholder \
    AWS_REGION=ap-southeast-1 \
    BUCKET=placeholder \
    BUCKET_PATH=uploads \
    AUTH_TOKEN=auth_token \
    HQ_TOKEN=hq_token \
    HOST="http://localhost"

WORKDIR /app

COPY --from=build-env /app/server .

EXPOSE 8080
ENTRYPOINT /app/server
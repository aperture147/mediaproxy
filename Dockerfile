FROM golang:1.15-buster AS build-env

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=1.0.0 -w -s" -o server

FROM debian:buster-slim

LABEL maintainer="HaiVQ <haivq@house3d.net>"

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates ffmpeg && \
    rm -rf /var/lib/apt/lists/*

ENV AWS_ACCESS_KEY_ID=placeholder \
    AWS_SECRET_ACCESS_KEY=placeholder \
    AWS_REGION=ap-southeast-1 \
    AWS_S3_BUCKET=placeholder \
    AWS_S3_BUCKET_PATH=uploads \
    AUTH_TOKEN=auth_token \
    SPECIAL_AUTH_TOKEN=special_auth_token \
    HOST="http://localhost"

WORKDIR /app

COPY --from=build-env /app/server .

EXPOSE 8080
ENTRYPOINT /app/server
# ビルド用イメージ
FROM golang:1.19.2-alpine3.16 AS builder

RUN apk update && apk add --no-cache git
RUN apk add --no-cache alpine-sdk build-base

WORKDIR /api

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/cosmtrek/air@latest

COPY . .
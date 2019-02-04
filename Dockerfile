ARG BASE=alpine

FROM golang:1.11.2-alpine AS builder
RUN apk add --no-cache git

RUN mkdir -p /src
WORKDIR /src/

ADD clip /src/clip
ADD hue /src/hue
ADD go.mod go.sum onebridge.go /src/

ARG GOOS=linux
RUN echo "$GOOS"
ARG GOARCH=amd64
RUN echo "$GOARCH"
RUN env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o /main onebridge.go

FROM $BASE
ARG BASE=alpine
RUN echo "$BASE"

RUN mkdir /app
WORKDIR /app
COPY debug /app/debug
COPY docs /app/docs
COPY --from=builder /main /
CMD ["/main", "-data", "/data"]

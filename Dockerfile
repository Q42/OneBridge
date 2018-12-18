ARG BASE=alpine

FROM golang:1.11 AS builder

RUN mkdir -p /go/src/onebridge
WORKDIR /go/src/onebridge

ADD clip /go/src/onebridge/clip
ADD hue /go/src/onebridge/hue
ADD onebridge.go /go/src/onebridge
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

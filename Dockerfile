FROM golang:1.9 AS builder

RUN mkdir -p /go/src/onebridge
WORKDIR /go/src/onebridge

RUN go get -u github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

ADD clip /go/src/onebridge/clip
ADD hue /go/src/onebridge/hue
ADD onebridge.go /go/src/onebridge
RUN env GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -o /main onebridge.go

FROM resin/raspberry-pi2-alpine
RUN mkdir /app
WORKDIR /app
COPY debug /app/debug
COPY docs /app/docs
COPY --from=builder /main /
CMD ["/main"]
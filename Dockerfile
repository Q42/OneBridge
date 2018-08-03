FROM golang:1.9 AS builder

RUN mkdir -p /go/src/OneBridge
WORKDIR /go/src/onebridge

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

ADD clip /go/src/onebridge/clip
ADD hue /go/src/onebridge/hue
ADD onebridge.go /go/src/onebridge
RUN CGO_ENABLED=0 go build -o /main onebridge.go

FROM alpine
RUN mkdir /app
WORKDIR /app
COPY debug /app/debug
COPY docs /app/docs
COPY --from=builder /main /
CMD ["/main"]
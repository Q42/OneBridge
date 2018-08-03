FROM golang:onbuild as gobuilder
RUN mkdir -p /go/src/github.com/Q42/OneBridge
ADD clip hue onebridge.go /go/src/github.com/Q42/OneBridge/
WORKDIR /go/src/github.com/Q42/OneBridge
RUN go build onebridge.go -o main .

FROM scratch
RUN mkdir /app
WORKDIR /app
COPY debug docs /app/
COPY --from=gobuilder main /
CMD ["/main"]
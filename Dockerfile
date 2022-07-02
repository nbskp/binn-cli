FROM golang:1.17.2-alpine3.14 AS builder
ENV GO111MODULE=off
RUN apk --update add alpine-sdk

WORKDIR /go/src/github.com/binn-client
COPY ./ ./

RUN CGO_ENABLED=0 go build -o /go/bin/binn main.go

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/binn /binn
CMD ["/binn"]

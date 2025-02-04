FROM golang:1.23.5-alpine3.21 AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/app-pinger/pinger

RUN apk --update --no-cache add ca-certificates gcc libtool make musl-dev protoc git

COPY pinger /go/src/app-pinger/pinger
COPY pkg /go/src/app-pinger/pkg
COPY go.mod /go/src/app-pinger
COPY go.sum /go/src/app-pinger
COPY .env /go/src/app-pinger

RUN go mod download

RUN go build -o /pinger ./main.go

FROM scratch

COPY --from=builder /pinger pinger
COPY --from=builder /go/src/app-pinger/.env ./

EXPOSE 8080

ENTRYPOINT ["/pinger"]
FROM golang:1.23.5-alpine3.21 AS builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/app-pinger/backend

RUN apk --update --no-cache add ca-certificates gcc libtool make musl-dev protoc git

COPY backend /go/src/app-pinger/backend
COPY pkg /go/src/app-pinger/pkg
COPY go.mod /go/src/app-pinger
COPY go.sum /go/src/app-pinger
COPY .env /go/src/app-pinger

RUN go mod download

RUN go build -o /backend ./cmd/backend/main.go

FROM scratch

COPY --from=builder /backend backend
COPY --from=builder /go/src/app-pinger/backend/internal/migrations ./migrations
COPY --from=builder /go/src/app-pinger/.env ./
COPY --from=builder /go/src/app-pinger/backend/config/verifier_config.yaml ./

EXPOSE 8082

ENTRYPOINT ["/backend"]
# Dockerfile.distroless
FROM golang:1.21-alpine as base

ENV APP_HOME /go/src/blog
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"

COPY . .

RUN go mod download
RUN go mod verify
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main .

FROM gcr.io/distroless/static-debian11 as production

COPY --from=base /main .

CMD ["./main", "--listen-addr", ":22222"]

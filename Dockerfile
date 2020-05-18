FROM golang:1.14 as build-env

WORKDIR /app
COPY . /app

ENV GO111MODULE=on

RUN go build -o /app/app

FROM gcr.io/distroless/base
COPY --from=build-env /app/app /
ENTRYPOINT ["/app"]

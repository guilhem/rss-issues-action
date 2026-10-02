FROM golang:1.27.1 AS build-env

ENV CGO_ENABLED=0

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -trimpath -ldflags="-s -w" -o /bin/app .

FROM gcr.io/distroless/static-debian13

# GitHub mounts environment files owned by the runner; the action must write them.
USER 0:0

COPY --from=build-env /bin/app /

ENTRYPOINT ["/app"]

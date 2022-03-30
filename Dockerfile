FROM golang:1.18 as build-env


ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN apt-get -qq update && \
  apt-get -yqq install upx

WORKDIR /src
COPY . .

RUN go build \
  -a \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -tags netgo \
  -o /bin/app \
  . \
  && strip /bin/app \
  && upx -q -9 /bin/app

FROM gcr.io/distroless/static

COPY --from=build-env /bin/app /

ENTRYPOINT ["/app"]

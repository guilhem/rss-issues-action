FROM golang:1.21-alpine AS build-env


ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

# Install upx
RUN apk add --no-cache upx
RUN apk add --no-cache binutils

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

FROM gcr.io/distroless/base

# Error when writing in Environment Files
#USER nobody:nobody

COPY --from=build-env /bin/app /

ENTRYPOINT ["/app"]

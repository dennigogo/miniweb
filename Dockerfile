FROM golang:alpine AS build-env

RUN apk add make git

ADD . /src
RUN cd /src && make build

FROM alpine
WORKDIR /app
COPY --from=build-env /src/miniweb /app/
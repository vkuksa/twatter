# syntax=docker/dockerfile:1

FROM golang:1.20.4-alpine3.17 AS build-stage

WORKDIR /build/

COPY ./cmd/spammer/ . 

RUN CGO_ENABLED=0 GOOS=linux go build -o ./spammer ./main.go

#-

FROM scratch AS build-release-stage

WORKDIR /

COPY --from=build-stage /build/spammer .

EXPOSE 8080

ENTRYPOINT ["./spammer"]

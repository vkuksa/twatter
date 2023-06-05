# syntax=docker/dockerfile:1

FROM golang:1.20.4-alpine3.17 AS build-stage

WORKDIR /build/

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ./twatterd ./cmd/twatterd/main.go

#-

FROM scratch AS build-release-stage

WORKDIR /

COPY --from=build-stage /build/twatterd .

EXPOSE 8080

ENTRYPOINT ["./twatterd"]

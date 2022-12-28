FROM golang:1.16 AS build

WORKDIR /app/eslib
ADD https://github.com/masterofless/tickey-lickey/tree/main/apps/eslib/eslib.go ./
ADD https://github.com/masterofless/tickey-lickey/tree/main/apps/eslib/go.mod ./

WORKDIR /app/rmqlib
ADD https://github.com/masterofless/tickey-lickey/tree/main/apps/rmqlib/rmqlib.go ./
ADD https://github.com/masterofless/tickey-lickey/tree/main/apps/rmqlib/go.mod ./

WORKDIR /app/redeemer
COPY go.* ./
COPY *.go ./
RUN go mod download

RUN go build -o redeemer
RUN go test

FROM almalinux:latest

RUN groupadd notroot
RUN useradd notroot -g notroot

WORKDIR /app
COPY --from=build /app/redeemer/redeemer /app/redeemer
USER notroot:notroot

EXPOSE 8080

CMD [ "/app/redeemer" ]

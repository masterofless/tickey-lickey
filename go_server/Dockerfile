FROM golang:1.16 AS build

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go test
RUN go build -o /runtime

FROM alpine:latest

RUN addgroup notroot
RUN adduser -G notroot -D notroot
USER notroot:notroot

WORKDIR /
COPY --from=build /runtime /runtime

EXPOSE 8080

CMD [ "/runtime" ]
